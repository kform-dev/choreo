/*
Copyright 2024 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package recrunner

import (
	"context"
	"errors"
	"fmt"
	"sync"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/controller/collector"
	"github.com/kform-dev/choreo/pkg/controller/collector/result"
	"github.com/kform-dev/choreo/pkg/controller/informers"
	"github.com/kform-dev/choreo/pkg/controller/reconciler"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"github.com/kform-dev/choreo/pkg/util/inventory"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

type Runner interface {
	Start(ctx context.Context, branch string) error
	RunOnce(ctx context.Context, branch string) (*runnerpb.Once_Response, error)
	Stop()
	AddDiscoveryClient(discovery.CachedDiscoveryInterface)
	AddResourceClient(resourceclient.Client)
	AddContext(ctx context.Context)
}

func New(flags *genericclioptions.ConfigFlags, choreo choreo.Choreo) Runner {
	return &runner{
		flags:  flags,
		choreo: choreo,
	}
}

type runner struct {
	m      sync.RWMutex
	status RunnerStatus
	// added on new
	flags  *genericclioptions.ConfigFlags
	choreo choreo.Choreo
	// added dynamically before start
	ctx             context.Context
	discoveryClient discovery.CachedDiscoveryInterface
	client          resourceclient.Client
	// dynamic data
	cancel             context.CancelFunc
	reconcilerConfigs  []*choreov1alpha1.Reconciler
	libs               *unstructured.UnstructuredList
	reconcilerResultCh chan result.Result
	runResultCh        chan *runnerpb.Once_Response
	collector          collector.Collector
	informerfactory    informers.InformerFactory
}

func (r *runner) AddDiscoveryClient(discoveryClient discovery.CachedDiscoveryInterface) {
	r.discoveryClient = discoveryClient
}
func (r *runner) AddResourceClient(client resourceclient.Client) {
	r.client = client
}

func (r *runner) AddContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *runner) Start(ctx context.Context, branch string) error {
	if r.getStatus() != RunnerStatus_Stopped {
		return fmt.Errorf("runner is already running, status %s", r.status.String())
	}
	if err := r.runValidate(ctx, branch); err != nil {
		return err
	}

	// we use the server context to cancel/handle the status of the server
	// since the ctx we get is from the client
	runctx, cancel := context.WithCancel(r.ctx)
	r.setStatusAndCancel(RunnerStatus_Running, cancel)

	go func() {
		defer r.Stop()
		select {
		case <-runctx.Done():
			return
		default:
			// use runctx since the ctx is from the cmd and it will be cancelled upon completion
			r.runReconciler(runctx, branch, false) // false -> run continuously, not once
		}
	}()
	return nil
}

func (r *runner) RunOnce(ctx context.Context, branch string) (*runnerpb.Once_Response, error) {
	if r.getStatus() != RunnerStatus_Stopped {
		return nil, fmt.Errorf("runner is already running, status %s", r.status.String())
	}

	if err := r.runValidate(ctx, branch); err != nil {
		return nil, err
	}

	// we use the server context to cancel/handle the status of the server
	// since the ctx we get is from the client
	_, cancel := context.WithCancel(r.ctx)
	r.setStatusAndCancel(RunnerStatus_Once, cancel)

	defer r.Stop()

	return r.runReconciler(ctx, branch, true) // run once
}

func (r *runner) Stop() {
	if r.getCancel() != nil {
		r.cancel() // Cancel the context, which triggers stopping in the StartContinuous loop
	}
	r.setStatusAndCancel(RunnerStatus_Stopped, nil)
	// don't nilify the other resources, since they will be reinitialized
}

func (r *runner) runValidate(ctx context.Context, branch string) error {
	apiResources, err := r.discoveryClient.APIResources(ctx, branch)
	if err != nil {
		return err
	}
	var errm error
	if err := r.choreo.GetBranchStore().LoadData(ctx, branch); err != nil {
		errm = errors.Join(errm, err)
	}

	// we use this to garbagecollect - root object they might have gotten deleted
	inv := inventory.Inventory{}
	if err := inv.Build(ctx, r.client, r.discoveryClient, branch); err != nil {
		errm = errors.Join(errm, err)
	}
	garbageSet := inv.CollectGarbage()
	for _, ref := range garbageSet.UnsortedList() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind))
		u.SetName(ref.Name)
		u.SetNamespace(ref.Namespace)

		if err := r.client.Delete(ctx, u, &resourceclient.DeleteOptions{
			Branch: branch,
		}); err != nil {
			errm = errors.Join(errm, err)
		}
	}
	if errm != nil {
		return errm
	}

	reconcilerConfigs := &unstructured.UnstructuredList{}
	reconcilerConfigs.SetGroupVersionKind(choreov1alpha1.SchemeGroupVersion.WithKind(choreov1alpha1.ReconcilerKind))
	r.client.List(ctx, reconcilerConfigs, &resourceclient.ListOptions{
		ExprSelector: &resourcepb.ExpressionSelector{},
		Branch:       branch,
	})

	r.reconcilerConfigs = []*choreov1alpha1.Reconciler{}
	reconcilerGVKs := sets.New[schema.GroupVersionKind]()
	for _, recu := range reconcilerConfigs.Items {
		b, err := yaml.Marshal(recu.Object)
		if err != nil {
			errm = errors.Join(errm, err)
			continue
		}
		reconciler := choreov1alpha1.Reconciler{}
		if err := yaml.Unmarshal(b, &reconciler); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
		r.reconcilerConfigs = append(r.reconcilerConfigs, &reconciler)
		for _, gvk := range reconciler.GetGVKs().UnsortedList() {
			if !apiResources.Has(gvk) {
				errm = errors.Join(errm, fmt.Errorf("reconciler %s api %s of reconciler not available in apigroup", recu.GetName(), gvk.String()))
			}
		}
		reconcilerGVKs.Insert(reconciler.GetGVKs().UnsortedList()...)
	}

	libs := &unstructured.UnstructuredList{}
	libs.SetGroupVersionKind(choreov1alpha1.SchemeGroupVersion.WithKind(choreov1alpha1.LibraryKind))
	if err := r.client.List(ctx, libs, &resourceclient.ListOptions{
		ExprSelector: &resourcepb.ExpressionSelector{},
		Branch:       branch,
	}); err != nil {
		errm = errors.Join(errm, err)
	}
	r.libs = libs

	if errm != nil {
		return errm
	}

	r.reconcilerResultCh = make(chan result.Result)
	r.runResultCh = make(chan *runnerpb.Once_Response)
	r.collector = collector.New(r.reconcilerResultCh, r.runResultCh)
	r.informerfactory = informers.NewInformerFactory(r.client, reconcilerGVKs, branch)

	return errm
}

func (r *runner) runReconciler(ctx context.Context, branch string, once bool) (*runnerpb.Once_Response, error) {
	reconcilerfactory, err := reconciler.NewReconcilerFactory(
		ctx, r.client, r.informerfactory, r.reconcilerConfigs, r.libs, r.reconcilerResultCh, branch)

	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		go r.collector.Start(ctx, once)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		go reconcilerfactory.Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		go r.informerfactory.Start(ctx)
	}()

	select {
	case result, ok := <-r.runResultCh:
		if !ok {
			return nil, nil
		}
		return result, nil

	case <-ctx.Done():
		wg.Wait()
		return nil, nil
	}
}

func (r *runner) getStatus() RunnerStatus {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.status
}

func (r *runner) setStatusAndCancel(status RunnerStatus, cancel func()) {
	r.m.Lock()
	defer r.m.Unlock()
	r.status = status
	r.cancel = cancel
}

func (r *runner) getCancel() func() {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.cancel
}

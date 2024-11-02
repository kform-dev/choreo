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

package choreo

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/controller/collector"
	"github.com/kform-dev/choreo/pkg/controller/informers"
	"github.com/kform-dev/choreo/pkg/controller/reconciler"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/util/inventory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

type Runner interface {
	AddResourceClientAndContext(ctx context.Context, client resourceclient.Client)
	Start(ctx context.Context, bctx *BranchCtx) (*runnerpb.Start_Response, error)
	Stop()
	RunOnce(ctx context.Context, bctx *BranchCtx) (*runnerpb.Once_Response, error)
}

func NewRunner(choreo Choreo) Runner {
	return &run{
		choreo: choreo,
	}
}

type run struct {
	m      sync.RWMutex
	status RunnerStatus
	//
	choreo Choreo
	// added dynamically before start
	ctx    context.Context
	client resourceclient.Client
	//discoveryClient discovery.CachedDiscoveryInterface
	// dynamic data
	cancel             context.CancelFunc
	reconcilerConfigs  []*choreov1alpha1.Reconciler
	libs               *unstructured.UnstructuredList
	reconcilerResultCh chan *runnerpb.Result
	runResultCh        chan *runnerpb.Once_Response
	collector          collector.Collector
	informerfactory    informers.InformerFactory
}

func (r *run) AddResourceClientAndContext(ctx context.Context, client resourceclient.Client) {
	r.client = client
	r.ctx = ctx
}

func (r *run) Start(ctx context.Context, bctx *BranchCtx) (*runnerpb.Start_Response, error) {
	if r.getStatus() == RunnerStatus_Running {
		return &runnerpb.Start_Response{}, nil
	}
	if r.getStatus() == RunnerStatus_Once {
		return &runnerpb.Start_Response{},
			status.Errorf(codes.InvalidArgument, "runner is already running, status %s", r.status.String())
	}

	if err := r.load(ctx, bctx); err != nil {
		return &runnerpb.Start_Response{}, err
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
			r.runReconciler(runctx, bctx, false) // false -> run continuously, not once
		}
	}()
	return &runnerpb.Start_Response{}, nil
}

func (r *run) Stop() {
	if r.getCancel() != nil {
		r.cancel() // Cancel the context, which triggers stopping in the StartContinuous loop
	}
	r.setStatusAndCancel(RunnerStatus_Stopped, nil)
	// don't nilify the other resources, since they will be reinitialized
}

func (r *run) RunOnce(ctx context.Context, bctx *BranchCtx) (*runnerpb.Once_Response, error) {
	if r.getStatus() != RunnerStatus_Stopped {
		return &runnerpb.Once_Response{},
			status.Errorf(codes.InvalidArgument, "runner is already running, status %s", r.status.String())
	}

	if err := r.load(ctx, bctx); err != nil {
		return nil, err
	}

	// we use the server context to cancel/handle the status of the server
	// since the ctx we get is from the client
	_, cancel := context.WithCancel(r.ctx)
	r.setStatusAndCancel(RunnerStatus_Once, cancel)

	defer r.Stop()

	rsp, err := r.runReconciler(ctx, bctx, true) // run once
	if err != nil {
		return rsp, err
	}

	r.createSnapshot(ctx, bctx)

	return rsp, nil
}

func (r *run) createSnapshot(ctx context.Context, bctx *BranchCtx) error {
	uid := uuid.New().String()

	apiResources := bctx.APIStore.GetAPIResources()

	inv := inventory.Inventory{}
	if err := inv.Build(ctx, r.client, apiResources, &inventory.BuildOptions{
		ShowManagedField: true,
		Branch:           bctx.Branch,
		ShowChoreoAPIs:   true,
	}); err != nil {
		return status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	fmt.Println("create snapshot", uid)

	r.choreo.SnapshotManager().Create(uid, apiResources, inv)

	return nil
}

func (r *run) load(ctx context.Context, bctx *BranchCtx) error {
	// we only work with checkout branch
	apiResources := bctx.APIStore.GetAPIResources()

	if err := r.choreo.GetBranchStore().LoadData(ctx, bctx.Branch); err != nil {
		return status.Errorf(codes.Internal, "err: %s", err.Error())
	}

	// we use this to garbagecollect -
	// Root object might have been deleted in the input so we need to cleanup
	inv := inventory.Inventory{}
	if err := inv.Build(ctx, r.client, apiResources, &inventory.BuildOptions{
		ShowManagedField: true,
		Branch:           bctx.Branch,
		ShowChoreoAPIs:   false, // TO CHANGE
	}); err != nil {
		return status.Errorf(codes.Internal, "err: %s", err.Error())
	}

	var errm error
	garbageSet := inv.CollectGarbage()
	for _, ref := range garbageSet.UnsortedList() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind))
		u.SetName(ref.Name)
		u.SetNamespace(ref.Namespace)

		//fmt.Println("delete garbage", u.GetAPIVersion(), u.GetKind(), u.GetName())

		if err := r.client.Delete(ctx, u, &resourceclient.DeleteOptions{
			Branch: bctx.Branch,
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
		Branch:       bctx.Branch,
	})

	r.reconcilerConfigs = []*choreov1alpha1.Reconciler{}
	reconcilerGVKs := sets.New[schema.GroupVersionKind]()
	for _, reconcilerConfig := range reconcilerConfigs.Items {
		b, err := yaml.Marshal(reconcilerConfig.Object)
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
			if !HasAPIResource(apiResources, gvk) {
				errm = errors.Join(errm, fmt.Errorf("reconciler %s api %s of reconciler not available in apigroup", reconcilerConfig.GetName(), gvk.String()))
			}
		}
		reconcilerGVKs.Insert(reconciler.GetGVKs().UnsortedList()...)
	}

	libs := &unstructured.UnstructuredList{}
	libs.SetGroupVersionKind(choreov1alpha1.SchemeGroupVersion.WithKind(choreov1alpha1.LibraryKind))
	if err := r.client.List(ctx, libs, &resourceclient.ListOptions{
		ExprSelector: &resourcepb.ExpressionSelector{},
		Branch:       bctx.Branch,
	}); err != nil {
		errm = errors.Join(errm, err)
	}
	r.libs = libs

	if errm != nil {
		return errm
	}

	r.reconcilerResultCh = make(chan *runnerpb.Result)
	r.runResultCh = make(chan *runnerpb.Once_Response)
	r.collector = collector.New(r.reconcilerResultCh, r.runResultCh)
	r.informerfactory = informers.NewInformerFactory(r.client, reconcilerGVKs, bctx.Branch)

	return errm

}

func PrintAPIResource(apiResources []*discoverypb.APIResource) bool {
	for _, apiResource := range apiResources {
		gvk := schema.GroupVersionKind{
			Group:   apiResource.Group,
			Version: apiResource.Version,
			Kind:    apiResource.Kind,
		}
		fmt.Println("print apiresources", gvk.String())
	}
	return false
}

func HasAPIResource(apiResources []*discoverypb.APIResource, gvk schema.GroupVersionKind) bool {
	for _, apiResource := range apiResources {
		if apiResource.Group == gvk.Group && apiResource.Version == gvk.Version && apiResource.Kind == gvk.Kind {
			return true
		}
	}
	return false
}

func (r *run) runReconciler(ctx context.Context, bctx *BranchCtx, once bool) (*runnerpb.Once_Response, error) {
	reconcilerfactory, err := reconciler.NewReconcilerFactory(
		ctx, r.client, r.informerfactory, r.reconcilerConfigs, r.libs, r.reconcilerResultCh, bctx.Branch)

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

func (r *run) getStatus() RunnerStatus {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.status
}

func (r *run) setStatusAndCancel(status RunnerStatus, cancel func()) {
	r.m.Lock()
	defer r.m.Unlock()
	r.status = status
	r.cancel = cancel
}

func (r *run) getCancel() func() {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.cancel
}

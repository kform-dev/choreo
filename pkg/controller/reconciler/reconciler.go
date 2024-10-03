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

package reconciler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/henderiw/logger/log"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	reconcileresult "github.com/kform-dev/choreo/pkg/controller/collector/result"
	"github.com/kform-dev/choreo/pkg/controller/eventhandler"
	"github.com/kform-dev/choreo/pkg/controller/informers"
	"github.com/kform-dev/choreo/pkg/controller/reconcile"
	"github.com/kform-dev/choreo/pkg/controller/reconciler/gotemplate"
	"github.com/kform-dev/choreo/pkg/controller/reconciler/starlark"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/util/workqueue"
)

type Reconciler interface {
	start(ctx context.Context)
}

func newReconciler(
	ctx context.Context,
	name string,
	client resourceclient.Client,
	informerFactory informers.InformerFactory,
	reconcilerConfig *choreov1alpha1.Reconciler,
	libs *unstructured.UnstructuredList,
	resultCh chan reconcileresult.Result,
	branchName string,
) Reconciler {
	r := &reconciler{
		name:                    name,
		maxConcurrentReconciles: 10,
		client:                  client,
		resultCh:                resultCh,
		branchName:              branchName,
	}

	r.createWorkQueue(ctx)
	r.addEventHandlerToInformerFactory(reconcilerConfig.DeepCopy(), informerFactory)

	var err error
	r.typedReconcilerFn, err = getTypeReconcilerFn(reconcilerConfig, libs, client, branchName)
	if err != nil {
		panic(err)
	}

	// need to create a consumer
	return r
}

type reconciler struct {
	name                    string
	queue                   workqueue.TypedRateLimitingInterface[types.NamespacedName]
	forgvk                  schema.GroupVersionKind
	maxConcurrentReconciles int
	typedReconcilerFn       reconcile.TypedReconcilerFn
	client                  resourceclient.Client
	resultCh                chan reconcileresult.Result
	branchName              string

	// Reconciler is a function that can be called at any time with the Name / Namespace of an object and
	// ensures that the state of the system matches the state specified in the object.
	// Defaults to the DefaultReconcileFunc.
	Do reconcile.TypedReconciler
}

func (r *reconciler) createWorkQueue(ctx context.Context) {
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[types.NamespacedName]()
	queue := workqueue.NewTypedRateLimitingQueueWithConfig(
		rateLimiter,
		workqueue.TypedRateLimitingQueueConfig[types.NamespacedName]{
			Name: r.name,
		})

	go func() {
		<-ctx.Done()
		queue.ShutDown()
	}()

	r.queue = queue
}

func (r *reconciler) addEventHandlerToInformerFactory(reconcilerConfig *choreov1alpha1.Reconciler, informerFactory informers.InformerFactory) {
	r.forgvk = reconcilerConfig.Spec.For.ResourceGVK.GetGVK()
	r.registerForResource(reconcilerConfig.Spec.For, informerFactory)
	for _, resource := range reconcilerConfig.Spec.Owns {
		if resource != nil {
			r.registerOwnResource(*resource, informerFactory)
		}
	}
	for _, resource := range reconcilerConfig.Spec.Watches {
		if resource != nil {
			r.registerWatchResource(*resource, informerFactory, reconcilerConfig.Spec.For)
		}
	}
}

func (r *reconciler) registerForResource(resource choreov1alpha1.ReconcilerResource, informerFactory informers.InformerFactory) {
	// this was validated before
	selector, _ := selector.ExprSelectorAsSelector(resource.Selector)
	eh := eventhandler.For{
		Name:     r.name,
		Queue:    r.queue,
		Client:   r.client,
		Selector: selector,
	}

	informerFactory.AddEventHandler(resource.ResourceGVK.GetGVK(), r.name, eh.EventHandler)
}

func (r *reconciler) registerOwnResource(resource choreov1alpha1.ReconcilerResource, informerFactory informers.InformerFactory) {
	eh := eventhandler.Own{
		Name:   r.name,
		Client: r.client,
		Queue:  r.queue,
		GVK:    r.forgvk, // for gvk
	}
	informerFactory.AddEventHandler(resource.ResourceGVK.GetGVK(), r.name, eh.EventHandler)
}

func (r *reconciler) registerWatchResource(resource choreov1alpha1.ReconcilerResource, informerFactory informers.InformerFactory, forResource choreov1alpha1.ReconcilerResource) {
	// we ignore the error since it was validated before
	watchselector, _ := selector.SelectorAsSelectorBuilder(resource.Selector)
	forSelector, _ := selector.ExprSelectorAsSelector(forResource.Selector)

	eh := eventhandler.Custom{
		Client:      r.client,
		Name:        r.name,
		Queue:       r.queue,
		Selector:    watchselector,
		ForGVK:      r.forgvk,
		ForSelector: forSelector,
		BranchName:  r.branchName,
	}

	informerFactory.AddEventHandler(resource.ResourceGVK.GetGVK(), r.name, eh.EventHandler)
}

func (r *reconciler) start(ctx context.Context) {
	// start the consumer
	log := log.FromContext(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(r.maxConcurrentReconciles)
	for i := 0; i < r.maxConcurrentReconciles; i++ {
		go func() {
			defer wg.Done()
			// Run a worker thread that just dequeues items, processes them, and marks them done.
			// It enforces that the reconcileHandler is never invoked concurrently with the same object.
			for r.processNextWorkItem(ctx) {
			}
		}()
	}
	<-ctx.Done()
	log.Debug("Shutdown signal received, waiting for all workers to finish")
	wg.Wait()
	log.Debug("All workers finished")
}

func (r *reconciler) processNextWorkItem(ctx context.Context) bool {
	req, shutdown := r.queue.Get()
	if shutdown {
		// Stop working
		return false
	}

	// We call Done here so the workqueue knows we have finished
	// processing this item. We also must remember to call Forget if we
	// do not want this work item being re-queued. For example, we do
	// not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off
	// period.
	defer r.queue.Done(req)

	r.reconcileHandler(ctx, req)
	return true
}

// Reconcile implements reconcile.Reconciler.
func (r *reconciler) Reconcile(ctx context.Context, req types.NamespacedName) (_ reconcile.Result, err error) {
	/*
		defer func() {
			if r := recover(); r != nil {
				if c.RecoverPanic != nil && *c.RecoverPanic {
					for _, fn := range utilruntime.PanicHandlers {
						fn(ctx, r)
					}
					err = fmt.Errorf("panic: %v [recovered]", r)
					return
				}

				log := logf.FromContext(ctx)
				log.Info(fmt.Sprintf("Observed a panic in reconciler: %v", r))
				panic(r)
			}
		}()
	*/
	reconciler, err := r.typedReconcilerFn()
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: req})
}

func (r *reconciler) initContext(ctx context.Context, req types.NamespacedName) context.Context {
	l := log.FromContext(ctx).With("reconciler", r.name, "req", req.String())
	return log.IntoContext(ctx, l)
}

func (r *reconciler) reconcileHandler(ctx context.Context, req types.NamespacedName) {
	ctx = r.initContext(ctx, req)
	log := log.FromContext(ctx)
	reconcileID := uuid.NewUUID()
	// TODO metrics
	reconcileRef := reconcileresult.ReconcileRef{
		ReconcilerName: r.name,
		GVK:            r.forgvk,
		Req:            req,
	}
	result := reconcileresult.Result{
		Operation:    runnerpb.Operation_START,
		ReconcileID:  reconcileID,
		ReconcileRef: reconcileRef,
		Time:         time.Now(),
	}

	r.resultCh <- result
	res, err := r.Reconcile(ctx, req)
	result.Time = time.Now()
	switch {
	case err != nil:
		//if errors.Is(err, reconcile.TerminalError(nil)) {
		//ctrlmetrics.TerminalReconcileErrors.WithLabelValues(c.Name).Inc()
		//} else {
		//r.queue.AddRateLimited(req)
		//}
		//ctrlmetrics.ReconcileErrors.WithLabelValues(c.Name).Inc()
		//ctrlmetrics.ReconcileTotal.WithLabelValues(c.Name, labelError).Inc()
		//if !result.IsZero() {
		//	log.Info("Warning: Reconciler returned both a non-zero result and a non-nil error. The result will always be ignored if the error is non-nil and the non-nil error causes reqeueuing with exponential backoff. For more details, see: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile#Reconciler")
		//}
		log.Error("reconcile error", "gvk", r.forgvk.String(), "req", req, "error", err)
		result.Operation = runnerpb.Operation_ERROR
		result.Message = err.Error()
		r.resultCh <- result
		r.queue.Forget(req)
	case res.RequeueAfter > 0:
		log.Debug("reconcile requeue", "after", res.RequeueAfter)
		// The result.RequeueAfter request will be lost, if it is returned
		// along with a non-nil error. But this is intended as
		// We need to drive to stable reconcile loops before queuing due
		// to result.RequestAfter
		result.Operation = runnerpb.Operation_REQUEUE
		r.resultCh <- result

		r.queue.Forget(req)
		r.queue.AddAfter(req, res.RequeueAfter)
		//ctrlmetrics.ReconcileTotal.WithLabelValues(c.Name, labelRequeueAfter).Inc()
	case res.Requeue:
		log.Debug("reconcile requeue")
		result.Operation = runnerpb.Operation_REQUEUE
		result.Message = res.Message
		r.resultCh <- result
		r.queue.AddRateLimited(req)
		//ctrlmetrics.ReconcileTotal.WithLabelValues(c.Name, labelRequeue).Inc()
	default:
		log.Debug("reconcile done")
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		result.Operation = runnerpb.Operation_STOP
		r.resultCh <- result
		r.queue.Forget(req)

		//ctrlmetrics.ReconcileTotal.WithLabelValues(c.Name, labelSuccess).Inc()
	}
}

func getTypeReconcilerFn(reconcilerConfig *choreov1alpha1.Reconciler, libs *unstructured.UnstructuredList, client resourceclient.Client, branch string) (reconcile.TypedReconcilerFn, error) {
	if reconcilerConfig.Spec.Type == nil {
		return nil, fmt.Errorf("reconcilerTypenot specified for %s", reconcilerConfig.GetName())
	}
	switch *reconcilerConfig.Spec.Type {
	case choreov1alpha1.SoftwardTechnologyType_Starlark:
		return starlark.NewReconcilerFn(client, reconcilerConfig, libs, branch), nil
	case choreov1alpha1.SoftwardTechnologyType_GoTemplate:
		return gotemplate.NewReconcilerFn(client, reconcilerConfig, branch), nil
	default:
		return nil, fmt.Errorf("reconcilerType %s is unsupported", (*reconcilerConfig.Spec.Type).String())
	}
}

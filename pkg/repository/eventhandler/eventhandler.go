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

package eventhandler

import (
	"context"
	"sync"

	"github.com/henderiw/logger/log"
	"k8s.io/client-go/util/workqueue"
)

type EventHandler interface {
	Handler()
	Start(ctx context.Context)
}

func NewEventHandler(ctx context.Context, name string, loader func(context.Context) error) EventHandler {
	return &apieh{
		queue:  createWorkQueue(ctx, name),
		loader: loader,
	}
}

type apieh struct {
	queue  workqueue.TypedRateLimitingInterface[struct{}]
	loader func(context.Context) error
}

func (r *apieh) Handler() {
	r.queue.Add(struct{}{})
}

func (r *apieh) Start(ctx context.Context) {
	log := log.FromContext(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < 1; i++ {
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

func (r *apieh) processNextWorkItem(ctx context.Context) bool {
	log := log.FromContext(ctx)
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

	if err := r.loader(ctx); err != nil {
		log.Error("dynamic apiloading failed", "error", err)
	}
	r.queue.Forget(req)

	return true
}

func createWorkQueue(ctx context.Context, name string) workqueue.TypedRateLimitingInterface[struct{}] {
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[struct{}]()
	queue := workqueue.NewTypedRateLimitingQueueWithConfig(
		rateLimiter,
		workqueue.TypedRateLimitingQueueConfig[struct{}]{
			Name: name,
		})

	go func() {
		<-ctx.Done()
		queue.ShutDown()
	}()

	return queue
}

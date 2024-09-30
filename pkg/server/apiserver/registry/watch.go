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

package registry

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watchermanager"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *storage) Watch(ctx context.Context, opts ...rest.ListOption) (watch.Interface, error) {
	o := rest.ListOptions{}
	o.ApplyOptions(opts)

	ctx, cancel := context.WithCancel(ctx)

	log := log.FromContext(ctx)
	log.Debug("watch")

	w := &watcher{
		cancel:         cancel,
		resultChan:     make(chan watch.Event),
		watcherManager: r.watcherManager,
		obj:            r.newFn(),
	}

	go w.listAndWatch(ctx, r, o)

	return w, nil
}

// implements the watchermanager Watcher interface
// implenents the k8s watch.Interface interface
type watcher struct {
	// interfce to the observer
	cancel         func()
	resultChan     chan watch.Event
	watcherManager watchermanager.WatcherManager
	obj            runtime.Unstructured

	// protection against concurrent access
	m             sync.Mutex
	eventCallback func(eventType resourcepb.Watch_EventType, obj runtime.Unstructured) bool
	done          bool
}

var _ watch.Interface = &watcher{}

// Stop stops watching. Will close the channel returned by ResultChan(). Releases
// any resources used by the watch.
func (r *watcher) Stop() {
	r.cancel()
}

// ResultChan returns a chan which will receive all the events. If an error occurs
// or Stop() is called, the implementation will close this channel and
// release any resources used by the watch.
func (r *watcher) ResultChan() <-chan watch.Event {
	return r.resultChan
}

// Implement the watcchermanafer.Watcher interface
// OnChange is the callback called when a object changes.
func (r *watcher) OnChange(eventType resourcepb.Watch_EventType, obj runtime.Unstructured) bool {
	r.m.Lock()
	defer r.m.Unlock()

	return r.eventCallback(eventType, obj)
}

func (r *watcher) listAndWatch(ctx context.Context, s *storage, options rest.ListOptions) {
	log := log.FromContext(ctx)
	if err := r.innerListAndWatch(ctx, s, options); err != nil {
		// TODO: We need to populate the object on this error
		// Most likely happens when we cancel a context, stop a watch
		log.Debug("sending error to watch stream", "error", err)
		ev := watch.Event{
			Type:   resourcepb.Watch_ERROR,
			Object: r.obj,
		}
		r.resultChan <- ev
	}

	log.Debug("stop listAndWatch")
	r.Stop()
}

// innerListAndWatch provides the callback handler
// 1. add a callback handler to receive any event we get while collecting the list of existing resources
// 2.
func (r *watcher) innerListAndWatch(ctx context.Context, s *storage, options rest.ListOptions) error {
	log := log.FromContext(ctx)

	errorResult := make(chan error)

	// backlog logs the events during startup
	var backlog []watch.Event
	// Make sure we hold the lock when setting the eventCallback, as it
	// will be read by other goroutines when events happen.
	r.m.Lock()
	r.eventCallback = func(eventType resourcepb.Watch_EventType, obj runtime.Unstructured) bool {
		if r.done {
			return false
		}
		backlog = append(backlog, watch.Event{
			Type:   eventType,
			Object: obj,
		})
		return true
	}
	r.m.Unlock()

	// we add the watcher to the watchermanager and start building a backlog for intermediate changes
	// while we startup, the backlog will be replayed once synced
	log.Debug("starting watch")
	if err := r.watcherManager.Add(ctx, options, r); err != nil {
		return err
	}

	// options.Watch means watch only no listing
	if !options.Watch {
		log.Debug("starting list watch")
		obj, err := s.List(ctx, &options)
		if err != nil {
			r.setDone()
			return err
		}

		var list unstructured.UnstructuredList
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &list); err != nil {
			r.setDone()
			return fmt.Errorf("expecting list type, got: %s, err: %s", reflect.TypeOf(obj).Name(), err.Error())
		}

		for _, obj := range list.Items {
			obj := obj
			ev := watch.Event{
				Type:   resourcepb.Watch_ADDED,
				Object: &obj,
			}
			log.Debug("listwatch", "obj", obj)
			r.sendWatchEvent(ctx, ev)
		}

		log.Debug("finished list watch")
	} else {
		log.Debug("watch only, no list")
	}

	// Repeatedly flush the backlog until we catch up
	for {
		r.m.Lock()
		chunk := backlog
		backlog = nil
		r.m.Unlock()

		if len(chunk) == 0 {
			break
		}
		log.Debug("flushing backlog", "chunk length", len(chunk))
		for _, ev := range chunk {
			r.sendWatchEvent(ctx, ev)
		}
	}

	r.m.Lock()
	// Pick up anything that squeezed in
	for _, ev := range backlog {
		r.sendWatchEvent(ctx, ev)
	}

	log.Debug("moving into streaming mode")
	r.eventCallback = func(eventType resourcepb.Watch_EventType, obj runtime.Unstructured) bool {
		accessor, _ := meta.Accessor(obj)
		log.Debug("eventCallBack", "eventType", eventType, "nsn", fmt.Sprintf("%s/%s", accessor.GetNamespace(), accessor.GetName()))
		if r.done {
			return false
		}
		ev := watch.Event{
			Type:   eventType,
			Object: obj,
		}
		r.sendWatchEvent(ctx, ev)
		return true
	}
	r.m.Unlock()

	select {
	case <-ctx.Done():
		r.setDone()
		return ctx.Err()

	case err := <-errorResult:
		r.setDone()
		return err
	}
}

func (r *watcher) sendWatchEvent(ctx context.Context, event watch.Event) {
	// TODO: Handle the case that the watch channel is full?
	if event.Object != nil {
		accessor, _ := meta.Accessor(event.Object)
		log := log.FromContext(ctx).With("event", event.Type, "nsn", fmt.Sprintf("%s/%s", accessor.GetNamespace(), accessor.GetName()))
		log.Debug("sending watch event")
	} else {
		log := log.FromContext(ctx).With("event", event.Type)
		log.Debug("sending watch event")
	}

	r.resultChan <- event
}

func (r *watcher) setDone() {
	r.m.Lock()
	defer r.m.Unlock()
	r.done = true
}

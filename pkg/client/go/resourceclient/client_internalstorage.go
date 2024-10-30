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

package resourceclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// used for api access during api loading
func NewAPIStorageClient(apistore *api.APIStore) Client {
	return &internal{
		apistore: apistore,
	}
}

type internal struct {
	apistore *api.APIStore
}

func (r *internal) Get(ctx context.Context, key types.NamespacedName, obj runtime.Unstructured, opts ...GetOption) error {
	o := GetOptions{}
	o.ApplyOptions(opts)

	storage, err := r.getStorage(&unstructured.Unstructured{Object: obj.UnstructuredContent()})
	if err != nil {
		return err
	}

	newobj, err := storage.Get(ctx, key.Name, &rest.GetOptions{
		Commit:            o.Commit,
		ShowManagedFields: o.ShowManagedFields,
		Trace:             o.Trace,
		Origin:            o.Origin,
	})
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(newobj.UnstructuredContent())
	return nil
}
func (r *internal) List(ctx context.Context, obj runtime.Unstructured, opts ...ListOption) error {
	o := ListOptions{}
	o.ApplyOptions(opts)

	selector, err := selector.ResourceExprSelectorAsSelector(o.ExprSelector)
	if err != nil {
		return err
	}

	storage, err := r.getStorage(&unstructured.Unstructured{Object: obj.UnstructuredContent()})
	if err != nil {
		return err
	}

	newobj, err := storage.List(ctx, &rest.ListOptions{
		Commit:            o.Commit,
		ShowManagedFields: o.ShowManagedFields,
		Trace:             o.Trace,
		Origin:            o.Origin,
		Selector:          selector,
		Watch:             false,
	})
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(newobj.UnstructuredContent())
	return nil
}
func (r *internal) Apply(ctx context.Context, obj runtime.Unstructured, opts ...ApplyOption) error {
	o := ApplyOptions{}
	o.ApplyOptions(opts)

	storage, err := r.getStorage(&unstructured.Unstructured{Object: obj.UnstructuredContent()})
	if err != nil {
		return err
	}

	newobj, err := storage.Apply(ctx, obj, &rest.ApplyOptions{
		Trace:        o.Trace,
		Origin:       o.Origin,
		DryRun:       o.DryRun,
		FieldManager: o.FieldManager,
		Force:        o.Force,
	})
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(newobj.UnstructuredContent())
	return nil
}
func (r *internal) Create(ctx context.Context, obj runtime.Unstructured, opts ...CreateOption) error {
	o := CreateOptions{}
	o.ApplyOptions(opts)

	storage, err := r.getStorage(&unstructured.Unstructured{Object: obj.UnstructuredContent()})
	if err != nil {
		return err
	}

	newobj, err := storage.Create(ctx, obj, &rest.CreateOptions{
		Trace:  o.Trace,
		Origin: o.Origin,
		DryRun: o.DryRun,
	})
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(newobj.UnstructuredContent())
	return nil
}
func (r *internal) Update(ctx context.Context, obj runtime.Unstructured, opts ...UpdateOption) error {
	o := UpdateOptions{}
	o.ApplyOptions(opts)

	storage, err := r.getStorage(&unstructured.Unstructured{Object: obj.UnstructuredContent()})
	if err != nil {
		return err
	}

	newobj, err := storage.Update(ctx, obj, &rest.UpdateOptions{
		Trace:  o.Trace,
		Origin: o.Origin,
		DryRun: o.DryRun,
	})
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(newobj.UnstructuredContent())
	return nil
}
func (r *internal) Delete(ctx context.Context, obj runtime.Unstructured, opts ...DeleteOption) error {
	o := DeleteOptions{}
	o.ApplyOptions(opts)

	u := &unstructured.Unstructured{Object: obj.UnstructuredContent()}

	storage, err := r.getStorage(u)
	if err != nil {
		return err
	}

	if _, err := storage.Delete(ctx, u.GetName(), &rest.DeleteOptions{
		Trace:  o.Trace,
		Origin: o.Origin,
		DryRun: o.DryRun,
	}); err != nil {
		return err
	}
	return nil
}
func (r *internal) Watch(ctx context.Context, obj runtime.Unstructured, opts ...ListOption) chan *resourcepb.Watch_Response {
	o := ListOptions{}
	o.ApplyOptions(opts)

	storage, err := r.getStorage(&unstructured.Unstructured{Object: obj.UnstructuredContent()})
	if err != nil {
		return nil
	}

	lopts := &rest.ListOptions{
		Commit:            o.Commit,
		ShowManagedFields: o.ShowManagedFields,
		Trace:             o.Trace,
		Origin:            o.Origin,
		Selector:          selector.Everything(),
		Watch:             o.Watch,
	}
	wi, err := storage.Watch(ctx, lopts)
	if err != nil {
		return nil
	}

	rspch := make(chan *resourcepb.Watch_Response)
	go r.watch(ctx, wi, rspch)
	return rspch
}

func (r *internal) watch(ctx context.Context, wi watch.Interface, rspch chan *resourcepb.Watch_Response) {
	log := log.FromContext(ctx)
	for {
		//defer close(rspch)
		select {
		case <-ctx.Done():
			log.Debug("grpc watch stopped, stopping storage watch")
			wi.Stop()
			return
		case watchEvent, ok := <-wi.ResultChan():
			if !ok {
				log.Debug("result channel closed, stopping storage watch")
				continue
			}

			if watchEvent.Type == resourcepb.Watch_ERROR {
				log.Error("received watch error", "event", watchEvent)
				continue
			}

			if watchEvent.Object == nil {
				log.Error("received nil object in watch event", "event", watchEvent)
				continue
			}

			// used for debug/display
			/*
				gvk := watchEvent.Object.GetObjectKind().GroupVersionKind()
				if !ok {
					log.Info("grpc watch done")
					return
				}
				log.Debug("grpc watch send event", "eventType", watchEvent.Type.String(), "gvk", gvk)
			*/
			b, err := json.Marshal(watchEvent.Object.UnstructuredContent())
			if err != nil {
				log.Error("grpc watch failed to marshal object", "error", err.Error())
				continue
			}

			rspch <- &resourcepb.Watch_Response{
				Object:    b,
				EventType: watchEvent.Type,
			}
		}
	}
}

func (r *internal) Close() error {
	return nil
}

func (r *internal) getStorage(u *unstructured.Unstructured) (rest.Storage, error) {
	gv, err := schema.ParseGroupVersion(u.GetAPIVersion())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid apiVersion, err: %s", err.Error())
	}
	gk := schema.GroupKind{Group: gv.Group, Kind: u.GetKind()}

	resctx, err := r.apistore.Get(gk)
	if err != nil {
		return nil, fmt.Errorf("gvk %s not found, %v", gk.String(), err)
	}

	return resctx.Storage, nil
}

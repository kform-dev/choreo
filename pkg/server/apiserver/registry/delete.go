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
	"time"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *storage) Delete(ctx context.Context, name string, opts ...rest.DeleteOption) (runtime.Unstructured, error) {
	o := rest.DeleteOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	log.Debug("delete")

	old, err := r.Get(ctx, name, &rest.GetOptions{
		ShowManagedFields: true,
		Trace:             "delete",
	})
	if err != nil {
		return nil, err // apierror context is already added
	}

	oldObjectMeta, err := meta.Accessor(old)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot access objectMeta err: %s", err.Error())
	}

	if err := r.deleteStrategy.PrepareForDelete(ctx, old); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation delete preparation failed: %v", err)
	}

	pendingFinalizers := len(oldObjectMeta.GetFinalizers()) != 0
	if pendingFinalizers {
		// update the deletion timestamp if it was not already set
		existingDeletionTimestamp := oldObjectMeta.GetDeletionTimestamp()
		now := time.Now()
		if existingDeletionTimestamp == nil || existingDeletionTimestamp.After(now) {
			newObj := old.DeepCopyObject()
			newObjectMeta, err := meta.Accessor(newObj)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "cannot access objectMeta err: %s", err.Error())
			}
			metaNow := metav1.NewTime(now)
			newObjectMeta.SetDeletionTimestamp(&metaNow)

			/*
				if err = r.storage.Update(store.ToKey(objectMeta.GetName()), obj); err != nil {
					return nil, status.Errorf(codes.Internal, "cannot update object in store, err: %s", err.Error())
				}

				r.notifyWatcher(ctx, watch.Event{
					Type:   resourcepb.Watch_MODIFIED,
					Object: obj,
				})
			*/
			new, ok := newObj.(runtime.Unstructured)
			if !ok {
				return nil, status.Errorf(codes.Internal, "fieldmanager does not return an unstructured object")
			}
			obj, err := r.update(ctx, new, old, &rest.UpdateOptions{DryRun: o.DryRun})
			if err != nil {
				return nil, err
			}

			return obj, nil
		}
		// deletion timestamp was already set, so no need to update,
		// we wait till the finalizers are removed
		return old, nil
	}

	if err := r.storage.Delete(store.ToKey(name)); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete object err: %s", err.Error())
	}
	r.notifyWatcher(ctx, watch.Event{
		Type:   resourcepb.Watch_DELETED,
		Object: old,
	})
	return old, nil

}

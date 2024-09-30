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

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"github.com/kform-dev/choreo/pkg/util/object"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *storage) Update(ctx context.Context, new runtime.Unstructured, opts ...rest.UpdateOption) (runtime.Unstructured, error) {
	o := rest.UpdateOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	log.Debug("update")

	newObjectMeta, err := meta.Accessor(new)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot access objectMeta err: %s", err.Error())
	}

	old, err := r.Get(ctx, newObjectMeta.GetName(), &rest.GetOptions{ShowManagedFields: true})
	if err != nil {
		return nil, err // apierror context is already added
	}

	return r.update(ctx, new, old, &rest.UpdateOptions{DryRun: o.DryRun})

}

func (r *storage) update(ctx context.Context, new, old runtime.Unstructured, _ *rest.UpdateOptions) (runtime.Unstructured, error) {
	newObjectMeta, err := meta.Accessor(new)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot access objectMeta err: %s", err.Error())
	}

	oldObjectMeta, err := meta.Accessor(old)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot access objectMeta from original object err: %s", err.Error())
	}

	if errs := r.updateStrategy.ValidateUpdate(ctx, new, old); len(errs) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", errs)
	}

	if err := r.updateStrategy.PrepareForUpdate(ctx, new, old); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot prepare for update err: %s", err.Error())
	}

	// if the resourceVersion mismatch we dont allow the update to go through, since this can lead to
	// inconsistencies
	if newObjectMeta.GetResourceVersion() != oldObjectMeta.GetResourceVersion() {
		return nil, status.Errorf(codes.PermissionDenied, "conflict resource version mismatch %s/%s", newObjectMeta.GetResourceVersion(), oldObjectMeta.GetResourceVersion())
	}

	if newObjectMeta.GetDeletionTimestamp() != nil && len(newObjectMeta.GetFinalizers()) == 0 {

		if err := r.storage.Delete(store.ToKey(newObjectMeta.GetName())); err != nil {
			return nil, status.Errorf(codes.Internal, "cannot delete object err: %s", err.Error())
		}
		r.notifyWatcher(ctx, watch.Event{
			Type:   resourcepb.Watch_DELETED,
			Object: new,
		})
		// deleted
		return new, nil
	}

	// if there is no change we dont need to do anything
	copiedold := old.DeepCopyObject().(runtime.Unstructured)
	removeManagedFieldsFromUnstructured(copiedold)
	copiednew := new.DeepCopyObject().(runtime.Unstructured)
	removeManagedFieldsFromUnstructured(copiednew)
	if apiequality.Semantic.DeepEqual(copiednew, copiedold) {
		return new, nil
	}
	//fmt.Println("upgate, deepequal, old\n", copiedold)
	//fmt.Println("upgate, deepequal, new\n", copiedobj)
	UpdateResourceVersion(newObjectMeta, oldObjectMeta)

	// there is a change in the object, check if the spec is equal or not
	specEqual, err := isSpecEqual(old, new)
	if err != nil {
		return nil, err
	}
	if !specEqual {
		// we update the generation if the spec changed, otherwise we leave it untouched
		UpdateGeneration(newObjectMeta, object.GetGeneration(old))
	}

	if err = r.storage.Update(store.ToKey(newObjectMeta.GetName()), new); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update object in store, err: %s", err.Error())
	}

	r.notifyWatcher(ctx, watch.Event{
		Type:   resourcepb.Watch_MODIFIED,
		Object: new,
	})
	return new, nil
}

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
	log.Debug("update choreoapiserver")

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

func (r *storage) update(ctx context.Context, new, old runtime.Unstructured, o *rest.UpdateOptions) (runtime.Unstructured, error) {
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

	recursion := false
	if len(o.DryRun) == 1 && o.DryRun[0] == "recursion" {
		recursion = true
		o.DryRun = []string{}
	}

	rnew, rold, err := r.updateStrategy.InvokeUpdate(ctx, new, old, recursion)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot prepare for update err: %s", err.Error())
	}
	new = rnew.(runtime.Unstructured)
	old = rold.(runtime.Unstructured)

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
	object.RemoveManagedFieldsFromUnstructured(ctx, copiedold)
	copiednew := new.DeepCopyObject().(runtime.Unstructured)
	object.RemoveManagedFieldsFromUnstructured(ctx, copiednew)

	if apiequality.Semantic.DeepEqual(copiednew, copiedold) {
		return new, nil
	}
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

	if len(o.DryRun) > 0 {
		return new, nil
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

/*
func printAndConvertID(u *unstructured.Unstructured) {
	// Extracting the nested map
	spec, found, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil || !found {
		fmt.Printf("Error accessing spec: %v\n", err)
		return
	}

	// Retrieving the ID field
	idValue, found := spec["id"]
	if !found {
		fmt.Println("ID field not found")
		return
	}

	// Type checking and converting
	switch id := idValue.(type) {
	case float64:
		// Convert float64 to int64 if necessary
		fmt.Printf("ID as float64: %f, converting to int64: %d\n", id, int64(id))
	case int64:
		fmt.Printf("ID as int64: %d\n", id)
	default:
		fmt.Printf("ID is of a type I don't know how to handle: %T\n", id)
	}
}

func deepEqualWithDebug(a, b interface{}) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	return reflectDeepEqual(va, vb, "")
}

// Recursive reflection function
func reflectDeepEqual(va, vb reflect.Value, path string) bool {
	if va.Type() != vb.Type() {
		fmt.Printf("Type mismatch at %s: %v != %v\n", path, va.Type(), vb.Type())
		return false
	}

	switch va.Kind() {
	case reflect.Ptr, reflect.Interface:
		return reflectDeepEqual(va.Elem(), vb.Elem(), path)
	case reflect.Struct:
		for i := 0; i < va.NumField(); i++ {
			fieldName := va.Type().Field(i).Name
			if !reflectDeepEqual(va.Field(i), vb.Field(i), path+"."+fieldName) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if va.Len() != vb.Len() {
			fmt.Printf("Length mismatch at %s: %d != %d\n", path, va.Len(), vb.Len())
			return false
		}
		for i := 0; i < va.Len(); i++ {
			if !reflectDeepEqual(va.Index(i), vb.Index(i), fmt.Sprintf("%s[%d]", path, i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if va.Len() != vb.Len() {
			fmt.Printf("Map length mismatch at %s: %d != %d\n", path, va.Len(), vb.Len())
			return false
		}
		for _, key := range va.MapKeys() {
			if !reflectDeepEqual(va.MapIndex(key), vb.MapIndex(key), path+"["+fmt.Sprint(key)+"]") {
				return false
			}
		}
		return true
	default:
		if va.CanInterface() && vb.CanInterface() && !reflect.DeepEqual(va.Interface(), vb.Interface()) {
			fmt.Printf("Value mismatch at %s: %v != %v\n", path, va.Interface(), vb.Interface())
			return false
		}
		return true
	}
}
*/

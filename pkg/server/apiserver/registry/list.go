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

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *storage) List(ctx context.Context, opts ...rest.ListOption) (runtime.Unstructured, error) {
	//log := log.FromContext(ctx)
	o := rest.ListOptions{}
	o.ApplyOptions(opts)

	newListObj := r.newListFn()
	v, err := GetListPrt(newListObj)
	if err != nil {
		return nil, err
	}

	listFunc := func(key store.Key, obj runtime.Unstructured) {
		// we don't filter by default
		filter := false
		if o.Selector != nil {
			if !o.Selector.Matches(obj.UnstructuredContent()) {
				filter = true
			}
		}

		if !filter {
			if !o.ShowManagedFields {
				copiedObj := obj.DeepCopyObject().(runtime.Unstructured)
				removeManagedFieldsFromUnstructured(copiedObj)
				removeResourceVersionAndGenerationFromUnstructured(copiedObj)
				AppendItem(v, copiedObj)
			} else {
				AppendItem(v, obj)
			}

		}
	}

	r.storage.List(listFunc, &store.ListOptions{Commit: o.Commit})
	return newListObj, nil
}

func GetListPrt(listObj runtime.Object) (reflect.Value, error) {
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return reflect.Value{}, err
	}
	v, err := conversion.EnforcePtr(listPtr)
	if err != nil || v.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("need ptr to slice: %v", err)
	}
	return v, nil
}

func AppendItem(v reflect.Value, obj runtime.Object) {
	v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
}

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

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *storage) Get(ctx context.Context, name string, opts ...rest.GetOption) (runtime.Unstructured, error) {
	// key is the name of the object
	o := rest.GetOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx).With("name", name)
	log.Debug("get")

	obj, err := r.storage.Get(store.ToKey(name), &store.GetOptions{Commit: o.Commit})
	if err != nil {
		return obj, status.Errorf(codes.NotFound, "err: %s", err.Error())
	}

	if !o.ShowManagedFields {
		copiedObj := obj.DeepCopyObject().(runtime.Unstructured)
		removeManagedFieldsFromUnstructured(copiedObj)
		removeResourceVersionAndGenerationFromUnstructured(copiedObj)
		return copiedObj, nil
	}
	return obj, nil
}

func removeManagedFieldsFromUnstructured(obj runtime.Unstructured) {
	// Access the unstructured content
	unstructuredContent := obj.UnstructuredContent()

	// Access the metadata section
	metadata, found, err := unstructured.NestedMap(unstructuredContent, "metadata")
	if err != nil || !found {
		return
	}

	// Remove the managedFields key from metadata
	delete(metadata, "managedFields")

	// Set the updated metadata back to the unstructured content
	err = unstructured.SetNestedMap(unstructuredContent, metadata, "metadata")
	if err != nil {
		fmt.Println("Error setting metadata:", err)
	}
}

func removeResourceVersionAndGenerationFromUnstructured(obj runtime.Unstructured) {
	// Access the unstructured content
	unstructuredContent := obj.UnstructuredContent()

	// Access the metadata section
	metadata, found, err := unstructured.NestedMap(unstructuredContent, "metadata")
	if err != nil || !found {
		return
	}

	// Remove the resourceVersion key from metadata
	delete(metadata, "resourceVersion")
	delete(metadata, "generation")

	// Set the updated metadata back to the unstructured content
	err = unstructured.SetNestedMap(unstructuredContent, metadata, "metadata")
	if err != nil {
		fmt.Println("Error setting metadata:", err)
	}
}

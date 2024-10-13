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

package object

import (
	"context"

	"github.com/henderiw/logger/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func RemoveManagedFieldsFromUnstructured(ctx context.Context, obj runtime.Unstructured) {
	log := log.FromContext(ctx)
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
		log.Error("failed setting metadata", "error", err)
	}
}

func RemoveResourceVersionAndGenerationFromUnstructured(ctx context.Context, obj runtime.Unstructured) {
	log := log.FromContext(ctx)
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
		log.Error("failed setting metadata", "error", err)
	}
}

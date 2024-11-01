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

package resource

import (
	"bytes"
	"fmt"
)

var exampleManifest = `
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating1
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfigurationList
items:
- apiVersion: admissionregistration.k8s.io/v1
  kind: MutatingWebhookConfiguration
  metadata:
    name: mutating2
- apiVersion: admissionregistration.k8s.io/v1
  kind: MutatingWebhookConfiguration
  metadata:
    name: mutating3
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: validating1
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfigurationList
items:
- apiVersion: admissionregistration.k8s.io/v1
  kind: ValidatingWebhookConfiguration
  metadata:
    name: validating2
- apiVersion: admissionregistration.k8s.io/v1
  kind: ValidatingWebhookConfiguration
  metadata:
    name: validating3
---
apiVersion: v1
kind: List
items:
- apiVersion: admissionregistration.k8s.io/v1
  kind: MutatingWebhookConfiguration
  metadata:
    name: mutating4
- apiVersion: admissionregistration.k8s.io/v1
  kind: ValidatingWebhookConfiguration
  metadata:
    name: validating4
---
`

// ExampleNewLocalBuilderLoad demonstrates using a local resource builder to read typed resources from a manifest
func ExampleNewLocalBuilder() {
	// Create a local builder...
	builder := NewLocalBuilder().
		Unstructured().
		// Provide input via a Reader.
		// As an alternative, could call Path(false, "/path/to/file") to read from a file.
		Stream(bytes.NewBufferString(exampleManifest), "input").
		// Flatten items contained in List objects
		Flatten().
		// Accumulate as many items as possible
		ContinueOnError()

	// Run the builder
	result := builder.Do()

	if err := result.err; err != nil {
		fmt.Println("builder error:", err)
		return
	}

	items, err := result.Infos()
	if err != nil {
		fmt.Println("infos error:", err)
		return
	}

	for _, item := range items {
		fmt.Printf("%s (%T)\n", item.String(), item.Object)
	}

	// Output:
	// Name: "mutating1", Namespace: "" (*unstructured.Unstructured)
	// Name: "mutating2", Namespace: "" (*unstructured.Unstructured)
	// Name: "mutating3", Namespace: "" (*unstructured.Unstructured)
	// Name: "validating1", Namespace: "" (*unstructured.Unstructured)
	// Name: "validating2", Namespace: "" (*unstructured.Unstructured)
	// Name: "validating3", Namespace: "" (*unstructured.Unstructured)
	// Name: "mutating4", Namespace: "" (*unstructured.Unstructured)
	// Name: "validating4", Namespace: "" (*unstructured.Unstructured)
}

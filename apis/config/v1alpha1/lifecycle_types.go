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

package v1alpha1

type Lifecycle struct {
	// DeletionPolicy specifies what will happen to the underlying resource
	// when this resource is deleted - either "delete" or "orphan" the
	// resource.
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty" protobuf:"bytes,1,opt,name=deletionPolicy,casttype=DeletionPolicy"`
}

// A DeletionPolicy determines what should happen to the underlying external
// resource when a managed resource is deleted.
type DeletionPolicy string

const (
	// DeletionOrphan means the external resource will orphaned when its managed
	// resource is deleted.
	DeletionOrphan DeletionPolicy = "orphan"

	// DeletionDelete means both the  external resource will be deleted when its
	// managed resource is deleted.
	DeletionDelete DeletionPolicy = "delete"
)

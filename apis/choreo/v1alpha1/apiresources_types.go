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

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIResourcesSpec defines the desired state of Reconciler
type APIResourcesSpec struct {
	// Groups define the resources per group
	Groups []*APIResourceGroup `json:"groups" protobuf:"bytes,1,opt,name=groups"`
}

type APIResourceGroup struct {
	// Resource defines name of the resourcce
	Resource string `json:"resource" protobuf:"bytes,1,opt,name=resource"`
	// Group defines the group of the resource
	Group string `json:"group" protobuf:"bytes,2,opt,name=group"`
	// Version defines the version of the resource
	Version string `json:"version" protobuf:"bytes,3,opt,name=version"`
	// Kind defines the kind of the resource
	Kind string `json:"kind" protobuf:"bytes,4,opt,name=kind"`
	// ListKind defines the kind of the resource
	ListKind string `json:"listKind" protobuf:"bytes,5,opt,name=listKind"`
	// Namespaced defines if the resource is namespaced
	Namespaced bool `json:"namespaced" protobuf:"bytes,6,opt,name=namespaced"`
	// Categories defines if the categories of the resource
	Categories []string `json:"categories,omitempty" protobuf:"bytes,7,opt,name=categories"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// APIResources defines the APIResources API
type APIResources struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec APIResourcesSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

var (
	APIResourcesKind = reflect.TypeOf(APIResources{}).Name()
)

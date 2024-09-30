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

// LibrarySpec defines the desired state of Library
type LibrarySpec struct {
	// Type defines the software technology this library contains
	Type SoftwardTechnologyType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`
	// Code supporting the Library
	Code string `json:"code" protobuf:"bytes,2,opt,name=code"`
}

// LibraryStatus defines the observed state of Library
type LibraryStatus struct {
	// ConditionedStatus provides the status of the Library using conditions
	//condition.ConditionedStatus `json:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:resource:scope=Cluster,categories={choreo}
// Library defines the Library API
type Library struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   LibrarySpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status LibraryStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true
// LibraryList contains a list of Librarys
type LibraryList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Library `json:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	LibraryKind = reflect.TypeOf(Library{}).Name()
)

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

	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkSpec struct {
	Name           string         `json:"name,omitempty" yaml:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Partition      string         `json:"partition,omitempty" yaml:"partition,omitempty" protobuf:"bytes,2,opt,name=partition"`
	Type           *NetworkType   `json:"type,omitempty" yaml:"type,omitempty" protobuf:"bytes,3,opt,name=type"`
	IPFamilyPolicy IPFamilyPolicy `json:"ipFamilyPolicy" yaml:"ipFamilyPolicy" protobuf:"bytes,4,opt,name=ipFamilyPolicy"`
}

type NetworkStatus struct {
	// ConditionedStatus provides the status of the resource using conditions
	// - a ready condition indicates the overall status of the resource
	condv1alpha1.ConditionedStatus `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
	// Name defines the name the provider assigned
	Name *string `json:"name,omitempty" yaml:"name,omitempty" protobuf:"bytes21,opt,name=name"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Network is the Schema for the Network API
type Network struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   NetworkSpec   `json:"spec,omitempty" yaml:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status NetworkStatus `json:"status,omitempty" yaml:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// NetworkList contains a list of ref.Config
type NetworkList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Network `json:"items" yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	NetworkKind     = reflect.TypeOf(Network{}).Name()
	NetworkKindList = reflect.TypeOf(NetworkList{}).Name()
)

/*
Copyright 2023 The Nephio Authors.

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
	kuidreqv1alpha1 "github.com/kform-dev/choreo/apis/kuid/req/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TBD how do we distinguish the loopback from the vnic(s)
type InterfaceSpec struct {
	// NetworkInstance defines the networkInstance to which this interface belongs
	// Name and optionally Namespace is used here
	Network string `json:"network" yaml:"network"`
	// AttachmentType defines if the interface is attached using a vlan or not
	// +kubebuilder:validation:Enum=none;vlan
	AttachmentType AttachmentType `json:"attachmentType,omitempty" yaml:"attachmentType,omitempty"`
	// IpFamilyPolicy defines the ip family policy on this interface to determine single stack, dual stack
	// +kubebuilder:validation:Enum=none;ipv6-only;ipv4-only;dual-stack
	IPFamilyPolicy kuidreqv1alpha1.IPFamilyPolicy `json:"ipFamilyPolicy,omitempty" yaml:"ipFamilyPolicy,omitempty"`
}

type InterfaceStatus struct {
	// ConditionedStatus provides the status of the IPClain using conditions
	// - a ready condition indicates the overall status of the resource
	condv1alpha1.ConditionedStatus `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
}

// +kubebuilder:object:root=true
type Interface struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   InterfaceSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status InterfaceStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// +kubebuilder:object:root=true
type InterfaceList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Interface `json:"items" yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

// Interface type metadata.
var (
	InterfaceKind     = reflect.TypeOf(Interface{}).Name()
	InterfaceListKind = reflect.TypeOf(InterfaceList{}).Name()
)

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CapacitySpec struct {
	// MaxUplinkThroughput defines the max uplink dataplane throughput
	MaxUplinkThroughput resource.Quantity `json:"maxUplinkThroughput,omitempty" yaml:"maxUplinkThroughput,omitempty"`
	// MaxDownlinkThroughput defines the max downlink dataplane throughput
	MaxDownlinkThroughput resource.Quantity `json:"maxDownlinkThroughput,omitempty" yaml:"maxDownlinkThroughput,omitempty"`
	// MaxSessions defines the max sessions of the control plane
	// expressed in unit of 1000s
	MaxSessions int `json:"maxSessions,omitempty" yaml:"maxSessions,omitempty"`
	// MaxSubscribers defines the max subscribers
	// expressed in unit of 1000s
	MaxSubscribers int `json:"maxSubscribers,omitempty" yaml:"maxSubscribers,omitempty"`
	// MaxNFConnections defines the max NF(s) that can be connected to this NF/device
	MaxNFConnections uint16 `json:"maxNFConnections,omitempty" yaml:"maxNFConnections,omitempty"`
}

type CapacityStatus struct {
	// ConditionedStatus provides the status of the IPClain using conditions
	// - a ready condition indicates the overall status of the resource
	condv1alpha1.ConditionedStatus `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
}

// +kubebuilder:object:root=true
type Capacity struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   CapacitySpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status CapacityStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// +kubebuilder:object:root=true
type CapacityList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Capacity `json:"items" yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

// Capacity type metadata.
var (
	CapacityKind     = reflect.TypeOf(Capacity{}).Name()
	CapacityListKind = reflect.TypeOf(CapacityList{}).Name()
)

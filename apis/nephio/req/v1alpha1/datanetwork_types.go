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

type DataNetworkSpec struct {
	// Pools defines the parameters of the IP pool associated with the DNN
	Pools []*Pool `json:"pools,omitempty"`
	// NetworkInstance defines the networkInstance context to which this DNN belongs
	// Name and optionally Namespace is used here
	Network string `json:"network" yaml:"network"`
}

type Pool struct {
	// Name defines the name of the pool
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// PrefixLength define the size of the pool
	PrefixLength uint32 `json:"prefixLength,omitempty" yaml:"prefixLength,omitempty"`
	// IpFamilyPolicy defines the ip family policy on this interface to determine single stack, dual stack
	// +kubebuilder:validation:Enum=none;ipv6-only;ipv4-only;dual-stack
	IPFamilyPolicy kuidreqv1alpha1.IPFamilyPolicy `json:"ipFamilyPolicy,omitempty" yaml:"ipFamilyPolicy,omitempty"`
}

type DataNetworkStatus struct {
	// ConditionedStatus provides the status of the IPClain using conditions
	// - a ready condition indicates the overall status of the resource
	condv1alpha1.ConditionedStatus `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
	// Pools contains the statuses of individual pools
	Pools []PoolStatus `yaml:"pools,omitempty" json:"pools,omitempty"`
}

type PoolStatus struct {
	// Name of the pool
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IPv4 defines the IPv4 status of the attachment
	IPv4 *PoolPrefixStatusIPv4 `json:"ipv4,omitempty" yaml:"ipv4,omitempty" protobuf:"bytes,2,opt,name=ipv4"`
	// IPv6 defines the IPv6 status of the attachment
	IPv6 *PoolPrefixStatusIPv6 `json:"ipv6,omitempty" yaml:"ipv6,omitempty" protobuf:"bytes,3,opt,name=prefix"`
}

type PoolPrefixStatusIPv4 struct {
	PoolStatusPrefixes `json:",inline" yaml:",inline"`
}

type PoolPrefixStatusIPv6 struct {
	PoolStatusPrefixes `json:",inline" yaml:",inline"`
}

type PoolStatusPrefixes struct {
	Prefixes []*PoolStatusPrefix `json:"prefixes,omitempty" yaml:"prefixes,omitempty" protobuf:"bytes,1,opt,name=prefix"`
}

type PoolStatusPrefix struct {
	// Prefix defines the prefix, claimed for the attachment
	// +optional
	Prefix *string `json:"prefix,omitempty" yaml:"prefix,omitempty" protobuf:"bytes,4,opt,name=prefix"`
}

// +kubebuilder:object:root=true
type DataNetwork struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   DataNetworkSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status DataNetworkStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// +kubebuilder:object:root=true
type DataNetworkList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []DataNetwork `json:"items" yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

// DataNetworkName type metadata.
var (
	DataNetworkKind     = reflect.TypeOf(DataNetwork{}).Name()
	DataNetworkListKind = reflect.TypeOf(DataNetworkList{}).Name()
)

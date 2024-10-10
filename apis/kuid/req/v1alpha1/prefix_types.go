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

	"github.com/henderiw/iputil"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	ipamv1alpha1 "github.com/kform-dev/choreo/apis/kuid/backend/ipam/v1alpha1"
	commonv1alpha1 "github.com/kuidio/kuid/apis/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PrefixSpec struct {
	Partition string `json:"partition,omitempty" yaml:"partition,omitempty" protobuf:"bytes,1,opt,name=partition"`
	Network   string `json:"network,omitempty" yaml:"network,omitempty" protobuf:"bytes,2,opt,name=network"`
	// PrefixType defines the prefixtype of the Prefix Request;
	// - network kind is used for physical, virtual nics on a device
	// - pool kind is used for allocating dedicated IP addresses
	// - aggregate kind is used for claiming an aggregate prefix; only used for networkInstance prefixes
	// +kubebuilder:validation:Enum=`network`;`aggregate`;`pool`;
	// +optional
	PrefixType *ipamv1alpha1.IPPrefixType `json:"prefixType,omitempty" yaml:"prefixType,omitempty" protobuf:"bytes,3,opt,name=prefixType"`
	// PrefixLength defines the prefix length for the IP Claim, Must be set when CreatePrefic is set
	// If not present we use assume /32 for ipv4 and /128 for ipv6
	// +optional
	PrefixLength *uint32 `json:"prefixLength,omitempty" yaml:"prefixLength,omitempty" protobuf:"varint,4,opt,name=prefixLength"`
	// AddressFamily defines the address family for the IP claim
	// +kubebuilder:validation:Enum=`ipv4`;`ipv6`
	// +kubebuilder:validation:Optional
	// +optional
	AddressFamily *iputil.AddressFamily `json:"addressFamily,omitempty" yaml:"addressFamily,omitempty" protobuf:"bytes,5,opt,name=addressFamily"`
	// ClaimLabels define the user defined labels and selector labels used
	// in resource claim
	commonv1alpha1.ClaimLabels `json:",inline" yaml:",inline" protobuf:"bytes,6,opt,name=claimLabels"`
}

type PrefixStatus struct {
	// ConditionedStatus provides the status of the IPClain using conditions
	// - a ready condition indicates the overall status of the resource
	condv1alpha1.ConditionedStatus `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
	// Prefix defines the claimed prefix
	// +optional
	Prefix *string `json:"prefix,omitempty" yaml:"prefix,omitempty" protobuf:"bytes,4,opt,name=prefix"`
	// DefaultGateway defines the default gateway IP for the claimed prefix
	// DefaultGateway is only relevant for prefix kind = network
	// +optional
	DefaultGateway *string `json:"defaultGateway,omitempty" yaml:"defaultGateway,omitempty" protobuf:"bytes,5,opt,name=defaultGateway"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Prefix is the Schema for the Prefix API
type Prefix struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   PrefixSpec   `json:"spec,omitempty" yaml:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status PrefixStatus `json:"status,omitempty" yaml:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// PrefixList contains a list of ref.Config
type PrefixList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Prefix `json:"items" yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	PrefixKind     = reflect.TypeOf(Prefix{}).Name()
	PrefixKindList = reflect.TypeOf(PrefixList{}).Name()
)

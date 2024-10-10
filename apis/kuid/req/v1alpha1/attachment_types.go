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
	idv1alpha1 "github.com/kform-dev/choreo/apis/kuid/id/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AttachmentSpec struct {
	idv1alpha1.PartitionAttachmentID `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`

	Network string `json:"network,omitempty" yaml:"network,omitempty" protobuf:"bytes,1,opt,name=network"`
	// IPFamilyPolicy defines the ip family policy on this interface to determine single stack, dual stack
	// +kubebuilder:validation:Enum=none;ipv6-only;ipv4-only;dual-stack
	IPFamilyPolicy IPFamilyPolicy `json:"ipFamilyPolicy,omitempty" yaml:"ipFamilyPolicy,omitempty" protobuf:"bytes,2,opt,name=vlanTagging"`
	// VLANTagging defines if the attchement requires VLANTagging
	VLANTagging *bool `json:"vlanTagging,omitempty" yaml:"vlanTagging,omitempty" protobuf:"bytes,3,opt,name=vlanTagging"`
}

type AttachmentStatus struct {
	// ConditionedStatus provides the status of the IPClain using conditions
	// - a ready condition indicates the overall status of the resource
	condv1alpha1.ConditionedStatus `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
	// IPv4 defines the IPv4 status of the attachment
	IPv4 *AttachmentStatusIPv4 `json:"ipv4,omitempty" yaml:"ipv4,omitempty" protobuf:"bytes,2,opt,name=ipv4"`
	// IPv6 defines the IPv6 status of the attachment
	IPv6 *AttachmentStatusIPv6 `json:"ipv6,omitempty" yaml:"ipv6,omitempty" protobuf:"bytes,3,opt,name=prefix"`
	// VLAN defines the VLAN status of the attachment
	VLAN *uint32 `json:"vlan,omitempty" yaml:"vlan,omitempty" protobuf:"bytes,3,opt,name=vlan"`
}

type AttachmentStatusIPv4 struct {
	AttachmentStatusPrefixes `json:",inline" yaml:",inline"`
}

type AttachmentStatusIPv6 struct {
	AttachmentStatusPrefixes `json:",inline" yaml:",inline"`
}

type AttachmentStatusPrefixes struct {
	Prefixes []*AttachmentStatusPrefix `json:"prefixes,omitempty" yaml:"prefixes,omitempty" protobuf:"bytes,1,opt,name=prefix"`
}

type AttachmentStatusPrefix struct {
	// Prefix defines the prefix, claimed for the attachment
	// +optional
	Prefix *string `json:"prefix,omitempty" yaml:"prefix,omitempty" protobuf:"bytes,4,opt,name=prefix"`
	// DefaultGateway defines the default gateway IP for the prefix
	// +optional
	DefaultGateway *string `json:"defaultGateway,omitempty" yaml:"defaultGateway,omitempty" protobuf:"bytes,5,opt,name=defaultGateway"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Attachment is the Schema for the Attachment API
type Attachment struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   AttachmentSpec   `json:"spec,omitempty" yaml:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status AttachmentStatus `json:"status,omitempty" yaml:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// AttachmentList contains a list of ref.Config
type AttachmentList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Attachment `json:"items" yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	AttachmentKind     = reflect.TypeOf(Attachment{}).Name()
	AttachmentKindList = reflect.TypeOf(AttachmentList{}).Name()
)

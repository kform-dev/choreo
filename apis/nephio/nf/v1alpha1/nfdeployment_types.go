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
	nephioreqv1alpha1 "github.com/nephio-project/api/nf_requirements/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NFDeploymentSpec defines the desired state of NFDeployment
type NFDeploymentSpec struct {
	Name string `json:"name" yaml:"name" protobuf:"bytes,1,opt,name=name"`
	// PartitionClusterID identifies the cluster identity this resource belongs to
	idv1alpha1.PartitionClusterID `json:",inline" yaml:",inline" protobuf:"bytes,2,opt,name=nodeID"`
	// Provider defines the provider implementing this resource.
	Provider string `json:"provider" yaml:"provider" protobuf:"bytes,3,opt,name=provider"`
	// capacity defines the capacity characteristics of the NF deployment
	// +optional
	Capacity *nephioreqv1alpha1.CapacitySpec `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	// Interfaces defines the interfaces associated with the NF deployment
	// +optional
	Interfaces []InterfaceConfig `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
	// NetworkInstances defines the network instances associated with the NF deployment
	// +optional
	NetworkInstances []NetworkInstance `json:"networkInstances,omitempty" yaml:"networkInstances,omitempty"`
	// ParametersRefs defines addiitonal KRM parameter references the nf depends upon
	// +optional
	ParametersRefs []*ObjectReference `json:"parametersRefs,omitempty" yaml:"parametersRefs,omitempty"`
}

type ObjectReference struct {
	// APIVersion of the target resources
	APIVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`

	// Kind of the target resources
	Kind string `yaml:"kind,omitempty" json:"kind,omitempty"`

	// Name of the target resource
	// +optional
	Name *string `yaml:"name" json:"name"`

	// Note: Namespace is not allowed; the namespace
	// must match the namespace of the PackageVariantSet resource
}

// InterfaceConfig defines the configuration of the interface
type InterfaceConfig struct {
	// Name defines the name of the interface
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name" yaml:"name"`
	// IPv4 defines the ipv4 configuration of the interface
	// +optional
	IPv4 *IPv4 `json:"ipv4,omitempty" yaml:"ipv4,omitempty"`
	// IPv6Config defines the ipv6 configuration of the interface
	// +optional
	IPv6 *IPv6 `json:"ipv6,omitempty" yaml:"ipv6,omitempty"`
	// VLANID defines the specific vlan id associated on this interface
	// +optional
	VLANID *uint16 `json:"vlanID,omitempty" yaml:"vlanID,omitempty"`
}

// IPv4 defines the configuration parameters of an ipv4 interface or peer
type IPv4 struct {
	// Address defines the IPv4 address and prefix length in CIDR notation
	// [IP prefix, range IPv4 with host bits]
	Address string `json:"address" yaml:"address"`
	// Gateway defines the IPv4 address associated to the interface as a gateway
	// +optional
	Gateway *string `json:"gateway,omitempty" yaml:"gateway,omitempty"`
}

// IPv6 defines the configuration parameters of an ipv6 interface or peer
type IPv6 struct {
	// Address defines the IPv6 address and prefix length in CIDR notation
	// [IP prefix, range IPv6 with host bits]
	Address string `json:"address" yaml:"address"`
	// Gateway defines the IPv6 address associated to the interface as a gateway
	// +optional
	Gateway *string `json:"gateway,omitempty" yaml:"gateway,omitempty"`
}

// A networkInstance is a Layer 3 forwarding construct
// such as a virtual routing and forwarding (VRF) instance,
type NetworkInstance struct {
	// Name defines the name of the network instance
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name" yaml:"name"`
	// interfaces defines the interfaces associated with the network instance
	// +optional
	Interfaces []string `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
	// Peers defines the peer configuration associated with the network instance
	// +optional
	//Peers []PeerConfig `json:"peers,omitempty" yaml:"peers,omitempty"`
	// DataNetworks defines the data networks assocated with the network instance
	// +optional
	//DataNetworks []DataNetwork `json:"dataNetworks,omitempty" yaml:"dataNetworks,omitempty"`
	// BGP defines the BGP configuration associated with the network instance
	// +optional
	//BGP *BGPConfig `json:"bgp,omitempty" yaml:"bgp,omitempty"`
}

// NFDeploymentStatus defines the observed state of NFDeployment
type NFDeploymentStatus struct {
	// ConditionedStatus provides the status of the IPClain using conditions
	// - a ready condition indicates the overall status of the resource
	condv1alpha1.ConditionedStatus `json:",inline" yaml:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// A NFDeployment represents a kubernetes NFDeployment
// +k8s:openapi-gen=true
type NFDeployment struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   NFDeploymentSpec   `json:"spec,omitempty" yaml:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status NFDeploymentStatus `json:"status,omitempty" yaml:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// NFDeploymentList contains a list of NFDeployments
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NFDeploymentList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []NFDeployment `json:"items" yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	NFDeploymentKind     = reflect.TypeOf(NFDeployment{}).Name()
	NFDeploymentKindList = reflect.TypeOf(NFDeploymentList{}).Name()
)

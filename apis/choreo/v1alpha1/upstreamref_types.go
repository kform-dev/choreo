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

type UpstreamRefType string

const (
	UpstreamRefType_CRD UpstreamRefType = "crd"
	UpstreamRefType_All UpstreamRefType = "all"
)

func (r UpstreamRefType) String() string {
	switch r {
	case UpstreamRefType_CRD:
		return "crd"
	case UpstreamRefType_All:
		return "all"
	default:
		return "invalid"
	}
}

// UpstreamRefSpec defines the desired state of the UpstreamRef
type UpstreamRefSpec struct {
	// Type defines the type of upstream ref
	// a api/crd type
	// a full type
	// +kubebuilder:validation:Enum=crd;all;
	// +kubebuilder:default:=full
	Type UpstreamRefType `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Priority defines the priority of the upstreamRef; used to define the sequence of execution
	// +kubebuilder:default:=10
	Priority int `json:"priority,omitempty" protobuf:"bytes,2,opt,name=priority"`
	// URL specifies the base URL for a given repository for example:
	//   `https://github.com/kubenet.dev/kubenet-catalog.git`
	URL string `json:"url" protobuf:"bytes,3,opt,name=url"`
	// Directory defines the name of the directory for the ref.
	// if not present the root directory is assumed
	Directory *string `json:"directory,omitempty" protobuf:"bytes,4,opt,name=directory"`
	// Ref defines the upstream reference
	Ref UpstreamReference `json:"ref" protobuf:"bytes,5,opt,name=ref"`
	// Credentials defines the name of the secret that holds the credentials to connect to the upstream Ref
	Credentials string `json:"credentials,omitempty" protobuf:"bytes,6,opt,name=credentials"`
	// Includes define the files to include
	// Typically used for CRD upstream types
	Includes []string `json:"includes,omitempty" protobuf:"bytes,7,opt,name=includes"`
}

type UpstreamReference struct {
	// +kubebuilder:validation:Enum=hash;tag;
	// +kubebuilder:default:=hash
	Type RefType `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Name defines the reference name
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={pkg, knet}
// ConfigGenerator defines the ConfigGenerator API
type UpstreamRef struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec UpstreamRefSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

var (
	UpstreamRefKind = reflect.TypeOf(UpstreamRef{}).Name()
)

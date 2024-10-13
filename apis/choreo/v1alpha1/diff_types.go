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

// DiffSpec defines the desired state of Diff
type DiffSpec struct {
}

// DiffStatus defines the observed state of Diff
type DiffStatus struct {
	Items []*DiffItem `json:"items" protobuf:"bytes,1,opt,name=items"`
}

type DiffitemStatus string

const (
	DiffitemStatus_Equal    DiffitemStatus = "equal"
	DiffitemStatus_Added    DiffitemStatus = "added"
	DiffitemStatus_Modified DiffitemStatus = "modified"
	DiffitemStatus_Deleted  DiffitemStatus = "deleted"
)

type DiffItem struct {
	// Resource defines the resource identifier on the basis of apiVersion (group/version) and kind
	ResourceGVK `json:",inline" protobuf:"bytes,1,opt,name=resource"`
	Name        string         `json:"name" protobuf:"bytes,2,opt,name=name"`
	Namespace   string         `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	Status      DiffitemStatus `json:"status" protobuf:"bytes,4,opt,name=status"`
	// Diff represents a diff string with the differences
	Diff *string `json:"diff,omitempty" protobuf:"bytes,5,opt,name=diff"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:resource:scope=Cluster,categories={choreo}
// Diff defines the Diff API
type Diff struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   DiffSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DiffStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true
// DiffList contains a list of Diffs
type DiffList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Diff `json:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	DiffKind = reflect.TypeOf(Diff{}).Name()
)

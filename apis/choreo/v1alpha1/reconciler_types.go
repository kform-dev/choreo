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

	selectorv1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReconcilerSpec defines the desired state of Reconciler
type ReconcilerSpec struct {
	// ConditionType defines the condition used by this reconciler to reflect the status of its operation
	ConditionType *string `json:"conditionType,omitempty" protobuf:"bytes,1,opt,name=conditionType"`
	// SpecUpdate indicates the reconciler is updating the spec with additional data
	SpecUpdate *bool `json:"specUpdate,omitempty" protobuf:"bytes,2,opt,name=specUpdate"`
	// For defines the resource and business logic of the reconciler for this Reconciler.
	For ReconcilerResource `json:"for" protobuf:"bytes,3,opt,name=for"`
	// Owns define the child resources this Reconciler generates as part of its business logic.
	// The For resource of this Reconciler owns the derived child resources.
	// The OwnerReferences are set by the internal reconciler logic. Changes to any of these resources
	// will trigger the Reconciler reconciler
	Owns []*ReconcilerResource `json:"owns,omitempty" protobuf:"bytes,4,rep,name=owns"`
	// Watches defines the resources on which the main reconciler can be retriggered. The pipeline/business logic
	// determines if the reconciler is to be retriggered.
	Watches []*ReconcilerResource `json:"watches,omitempty" protobuf:"bytes,5,opt,name=watches"`
	// Type defines the software technology this library contains
	Type *SoftwardTechnologyType `json:"type,omitempty" protobuf:"bytes,6,opt,name=type"`
	// Code supporting the reconciler
	Code map[string]string `json:"code,omitempty" protobuf:"bytes,7,rep,name=code"`
}

type ReconcilerResource struct {
	// Resource defines the resource identifier on the basis of apiVersion (group/version) and kind
	ResourceGVK `json:",inline" protobuf:"bytes,1,opt,name=resource"`
	// A selector to restrict the resources by their fields on which the pipeline is triggered
	// Defaults to everything.
	// +optional
	Selector *selectorv1alpha1.ExpressionSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`
}

type ResourceGVK struct {
	// Group defines the group and version of the resource
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Version defines the version of the resource
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// Kind defines the kind of the resource
	Kind string `json:"kind" protobuf:"bytes,3,opt,name=kind"`
}

// ReconcilerStatus defines the observed state of Reconciler
type ReconcilerStatus struct {
	// ConditionedStatus provides the status of the Reconciler using conditions
	//condition.ConditionedStatus `json:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:resource:scope=Cluster,categories={choreo}
// Reconciler defines the Reconciler API
type Reconciler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ReconcilerSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ReconcilerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ReconcilerList contains a list of Reconcilers
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ReconcilerList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Reconciler `json:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	ReconcilerKind = reflect.TypeOf(Reconciler{}).Name()
)

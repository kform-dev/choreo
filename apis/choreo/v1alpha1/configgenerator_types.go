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

// ConfigGeneratorSpec defines the desired state of ConfigGenerator
type ConfigGeneratorSpec struct {
	// Resource defines the resource identifier on the basis of apiVersion (group/version) and kind
	//ResourceGVK `json:",inline" protobuf:"bytes,1,opt,name=resource"`
	// ProviderSelector defines how to select the provider
	ProviderSelector ConfigGeneratorProviderSelector `json:"providerSelector" protobuf:"bytes,1,opt,name=providerSelector"`
	// Name of the basepath where the vendor translation logic is located
	// Right now we assume a structure <basePath>/<providername as per selector>/<group_resource>
	// This might be expanded in the future
	Dir *string `json:"dir,omitempty" protobuf:"bytes,2,opt,name=dir"`
	//Providers []*ConfigGeneratorProvider `json:"providers" protobuf:"bytes,3,rep,name=providers"`
}

/*
type ConfigGeneratorProvider struct {
	// Name of the basepath where the vendor translation logic is located
	// Right now we assume a structure <basePath>/<providername as per selector>/<group_resource>
	BasePath string `json:"basePath" protobuf:"bytes,2,opt,name=basePath"`
	// Name of the provider
	//Provider *string `json:"name" protobuf:"bytes,1,opt,name=name"`
	//File *string `json:"file" protobuf:"bytes,3,opt,name=file"`
}
*/

type ConfigGeneratorProviderSelector struct {
	// Resource defines the resource identifier on the basis of apiVersion (group/version) and kind
	ResourceGVK `json:",inline" protobuf:"bytes,1,opt,name=resource"`
	// match is a map of {expression,value} pairs. A single {expression,value} in the match
	// map is equivalent to an element of matchExpressions, whose expression field is "expression", the
	// operator is "In", and the values array contains only "value". The requirements are ANDed.
	// +optional
	Match map[string]string `json:"match" protobuf:"bytes,2,rep,name=match"`
	// FieldPath that provides the provider information
	FieldPath string `json:"fieldPath" protobuf:"bytes,3,opt,name=fieldPath"`
}

// ConfigGeneratorStatus defines the observed state of ConfigGenerator
type ConfigGeneratorStatus struct {
	// ConditionedStatus provides the status of the ConfigGenerator using conditions
	//condition.ConditionedStatus `json:",inline" protobuf:"bytes,1,opt,name=conditionedStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:resource:scope=Cluster,categories={pkg, knet}
// ConfigGenerator defines the ConfigGenerator API
type ConfigGenerator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ConfigGeneratorSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ConfigGeneratorStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true
// ConfigGeneratorList contains a list of ConfigGenerators
type ConfigGeneratorList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ConfigGenerator `json:"items" protobuf:"bytes,2,rep,name=items"`
}

var (
	ConfigGeneratorKind = reflect.TypeOf(ConfigGenerator{}).Name()
)

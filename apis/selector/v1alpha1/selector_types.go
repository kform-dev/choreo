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

// A expression selector is a query over a set of resources. The result of matchLabels and
// matchExpressions are ANDed.
// An empty selector matches all objects. A null selector matches no objects.
// +structType=atomic
type ExpressionSelector struct {
	// match is a map of {expression,value} pairs. A single {expression,value} in the match
	// map is equivalent to an element of matchExpressions, whose expression field is "expression", the
	// operator is "In", and the values array contains only "value". The requirements are ANDed.
	// +optional
	Match map[string]string `json:"match,omitempty" protobuf:"bytes,1,rep,name=match"`
	// matchExpressions is a list of field selector requirements. The requirements are ANDed.
	// +optional
	MatchExpressions []ExpressionSelectorRequirement `json:"matchExpressions,omitempty" protobuf:"bytes,2,rep,name=matchExpressions"`
}

// A expression selector requirement is a selector that contains values, an expression, and an operator that
// relates the expression and values.
type ExpressionSelectorRequirement struct {
	// expression is a cel expression
	// +patchMergeKey=expression
	// +patchStrategy=merge
	Expression string `json:"expression" patchStrategy:"merge" patchMergeKey:"expression" protobuf:"bytes,1,opt,name=expression"`
	// operator represents a expression's relationship to a set of values.
	Operator SelectorOperator `json:"operator" protobuf:"bytes,2,opt,name=operator"`
	// values is an array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty.
	// +optional
	Values []string `json:"values,omitempty" protobuf:"bytes,3,rep,name=values"`
}

// A label selector operator is the set of operators that can be used in a selector requirement.
type SelectorOperator string

const (
	SelectorOpEquals       SelectorOperator = "Equals"
	SelectorOpDoubleEquals SelectorOperator = "DoubleEquals"
	SelectorOpNotEquals    SelectorOperator = "NotEquals"
	SelectorOpIn           SelectorOperator = "In"
	SelectorOpNotIn        SelectorOperator = "NotIn"
	SelectorOpExists       SelectorOperator = "Exists"
	SelectorOpDoesNotExist SelectorOperator = "DoesNotExist"
	SelectorOpGreaterThan  SelectorOperator = "GreaterThan"
	SelectorOpLessThan     SelectorOperator = "LessThan"
)

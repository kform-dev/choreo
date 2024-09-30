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

package selector

import (
	selector1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
)

type Operator string

const (
	DoesNotExist Operator = "!"
	Equals       Operator = "="
	DoubleEquals Operator = "=="
	In           Operator = "in"
	NotEquals    Operator = "!="
	NotIn        Operator = "notin"
	Exists       Operator = "exists"
	GreaterThan  Operator = "gt"
	LessThan     Operator = "lt"
)

func toResourcePbOperator(op selector1alpha1.SelectorOperator) resourcepb.Operator {
	switch op {
	case selector1alpha1.SelectorOpEquals:
		return resourcepb.Operator_Equals
	case selector1alpha1.SelectorOpDoubleEquals:
		return resourcepb.Operator_DoubleEquals
	case selector1alpha1.SelectorOpNotEquals:
		return resourcepb.Operator_NotEquals
	case selector1alpha1.SelectorOpIn:
		return resourcepb.Operator_In
	case selector1alpha1.SelectorOpNotIn:
		return resourcepb.Operator_NotIn
	case selector1alpha1.SelectorOpExists:
		return resourcepb.Operator_Exists
	case selector1alpha1.SelectorOpDoesNotExist:
		return resourcepb.Operator_DoesNotExist
	case selector1alpha1.SelectorOpGreaterThan:
		return resourcepb.Operator_GreaterThan
	case selector1alpha1.SelectorOpLessThan:
		return resourcepb.Operator_LessThan
	default:
		// TBD check if we need to error
		return resourcepb.Operator_Equals
	}
}

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
	"errors"
	"fmt"
	"strconv"

	selector1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
)

func ResourceExprSelectorAsSelector(es *resourcepb.ExpressionSelector) (Selector, error) {
	if es == nil {
		return Nothing(), nil
	}
	if len(es.Match)+len(es.MatchExpressions) == 0 {
		return Everything(), nil
	}

	requirements := make([]Requirement, 0, len(es.Match)+len(es.MatchExpressions))
	for f, v := range es.Match {
		r, err := NewRequirement(f, resourcepb.Operator_Equals, []string{v})
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, r)
	}
	for _, expr := range es.MatchExpressions {
		r, err := NewRequirement(expr.Expression, expr.Operator, expr.Values)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, r)
	}
	selector := NewSelector()
	selector = selector.Add(requirements...)
	return selector, nil
}

func ExprSelectorAsSelector(es *selector1alpha1.ExpressionSelector) (Selector, error) {
	if es == nil {
		return Nothing(), nil
	}
	if len(es.Match)+len(es.MatchExpressions) == 0 {
		return Everything(), nil
	}

	requirements := make([]Requirement, 0, len(es.Match)+len(es.MatchExpressions))
	for f, v := range es.Match {
		r, err := NewRequirement(f, resourcepb.Operator_Equals, []string{v})
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, r)
	}
	for _, expr := range es.MatchExpressions {
		r, err := NewRequirement(expr.Expression, toResourcePbOperator(expr.Operator), expr.Values)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, r)
	}
	selector := NewSelector()
	selector = selector.Add(requirements...)
	return selector, nil
}

func SelectorAsSelectorBuilder(es *selector1alpha1.ExpressionSelector) (SelectorBuilder, error) {
	if es == nil {
		return nil, fmt.Errorf("a selector builder always needs a selector")
	}
	if len(es.Match) == 0 {
		return nil, fmt.Errorf("a selector builder always needs a match")
	}

	requirements := make([]BuilderRequirement, 0, len(es.Match)+len(es.MatchExpressions))
	for f, v := range es.Match {
		r, err := NewBuilderRequirement(f, resourcepb.Operator_Equals, []string{v})
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, r)
	}
	selectorbuilder := NewSelectorBuilder()
	selectorbuilder = selectorbuilder.Add(requirements...)
	return selectorbuilder, nil
}

// NewRequirement is the constructor for a Requirement.
// If any of these rules is violated, an error is returned:
//  1. The operator can only be In, NotIn, Equals, DoubleEquals, Gt, Lt, NotEquals, Exists, or DoesNotExist.
//  2. If the operator is In or NotIn, the values set must be non-empty.
//  3. If the operator is Equals, DoubleEquals, or NotEquals, the values set must contain one value.
//  4. If the operator is Exists or DoesNotExist, the value set must be empty.
//  5. If the operator is Gt or Lt, the values set must contain only one value, which will be interpreted as an integer.
//
// The empty string is a valid value in the input values set.
// Returned error, if not nil, is guaranteed to be an aggregated field.ErrorList
func NewRequirement(expr string, op resourcepb.Operator, vals []string) (Requirement, error) {
	var errm error
	switch op {
	case resourcepb.Operator_In, resourcepb.Operator_NotIn:
		if len(vals) == 0 {
			errm = errors.Join(errm, fmt.Errorf("for 'in', 'notin' operators, values set can't be empty"))
		}
	case resourcepb.Operator_Equals, resourcepb.Operator_DoubleEquals, resourcepb.Operator_NotEquals:
		if len(vals) != 1 {
			errm = errors.Join(errm, fmt.Errorf("exact-match compatibility requires one single value"))
		}
	case resourcepb.Operator_Exists, resourcepb.Operator_DoesNotExist:
		if len(vals) != 0 {
			errm = errors.Join(errm, fmt.Errorf("values set must be empty for exists and does not exist"))
		}
	case resourcepb.Operator_GreaterThan, resourcepb.Operator_LessThan:
		if len(vals) != 1 {
			errm = errors.Join(errm, fmt.Errorf("for 'Gt', 'Lt' operators, exactly one value is required"))
		}
		for i := range vals {
			if _, err := strconv.ParseInt(vals[i], 10, 64); err != nil {
				errm = errors.Join(errm, fmt.Errorf("for 'Gt', 'Lt' operators, the value must be an integer, got %v", vals[i]))
			}
		}
	default:
		errm = errors.Join(errm, fmt.Errorf("not a valid operator, got %v", op.String()))
	}
	return Requirement{&resourcepb.ExpressionSelectorRequirement{Expression: expr, Operator: op, Values: vals}}, errm
}

// NewRequirement is the constructor for a Requirement.
// If any of these rules is violated, an error is returned:
//  1. The operator can only be Equals, DoubleEquals, Gt, Lt, NotEquals, Exists, or DoesNotExist.
//  3. If the operator is Equals, DoubleEquals, or NotEquals, the values set must contain one value.
//
// The empty string is a valid value in the input values set.
// Returned error, if not nil, is guaranteed to be an aggregated field.ErrorList
func NewBuilderRequirement(expr string, op resourcepb.Operator, vals []string) (BuilderRequirement, error) {
	var errm error
	switch op {
	case resourcepb.Operator_Equals, resourcepb.Operator_DoubleEquals, resourcepb.Operator_NotEquals:
		if len(vals) != 1 {
			errm = errors.Join(errm, fmt.Errorf("exact-match compatibility requires one single value"))
		}
	default:
		errm = errors.Join(errm, fmt.Errorf("not a valid operator, got %v", op.String()))
	}
	return BuilderRequirement{&resourcepb.ExpressionSelectorRequirement{Expression: expr, Operator: op, Values: vals}}, errm
}

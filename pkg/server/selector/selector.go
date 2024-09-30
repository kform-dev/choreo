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
	"context"
	"strconv"
	"strings"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/selector/cel"
)

type Selector interface {
	// Matches returns true if this selector matches the given struct.
	Matches(map[string]any) bool

	// Add adds requirements to the Selector
	Add(r ...Requirement) Selector

	// String returns a human readable string that represents this selector.
	String() string
}

// Nothing returns a selector that matches no labels
func Nothing() Selector {
	return sharedNothingSelector
}

// Sharing this saves 1 alloc per use; this is safe because it's immutable.
var sharedNothingSelector Selector = nothingSelector{}

type nothingSelector struct{}

func (nothingSelector) Matches(map[string]any) bool     { return false }
func (n nothingSelector) Add(_ ...Requirement) Selector { return n }
func (n nothingSelector) String() string                { return "" }

// NewSelector returns a nil selector
func NewSelector() Selector {
	return internalSelector(nil)
}

// Everything returns a selector that matches all labels.
func Everything() Selector {
	return sharedEverythingSelector
}

// Sharing this saves 1 alloc per use; this is safe because it's immutable.
var sharedEverythingSelector Selector = internalSelector{}

type internalSelector Requirements

// Add adds requirements to the selector. It copies the current selector returning a new one
func (s internalSelector) Add(reqs ...Requirement) Selector {
	ret := make(internalSelector, 0, len(s)+len(reqs))
	ret = append(ret, s...)
	ret = append(ret, reqs...)
	//sort.Slice(func())
	return ret
}

// String returns a comma-separated string of all
// the internalSelector Requirements' human-readable strings.
func (s internalSelector) String() string {
	var reqs []string
	for ix := range s {
		reqs = append(reqs, s[ix].String())
	}
	return strings.Join(reqs, ",")
}

func (r internalSelector) Matches(data map[string]any) bool {
	for i := range r {
		if matches := r[i].Matches(data); !matches {
			return false
		}
	}
	return true
}

// Matches returns true if the Requirement matches the input Labels.
// There is a match in the following cases:
//  1. The operator is Exists and Labels has the Requirement's key.
//  2. The operator is In, Labels has the Requirement's key and Labels'
//     value for that key is in Requirement's value set.
//  3. The operator is NotIn, Labels has the Requirement's key and
//     Labels' value for that key is not in Requirement's value set.
//  4. The operator is DoesNotExist or NotIn and Labels does not have the
//     Requirement's key.
//  5. The operator is GreaterThanOperator or LessThanOperator, and Labels has
//     the Requirement's key and the corresponding value satisfies mathematical inequality.
func (r *Requirement) Matches(data map[string]any) bool {
	log := log.FromContext(context.Background())
	switch r.Operator {
	case resourcepb.Operator_In, resourcepb.Operator_Equals, resourcepb.Operator_DoubleEquals:
		v, found, err := cel.GetValue(data, r.Expression)
		if err != nil || !found {
			return false
		}
		return r.hasValue(v)
	case resourcepb.Operator_NotIn, resourcepb.Operator_NotEquals:
		v, found, err := cel.GetValue(data, r.Expression)
		if err != nil || !found {
			return true
		}
		return !r.hasValue(v)
	case resourcepb.Operator_Exists:
		_, found, err := cel.GetValue(data, r.Expression)
		if err != nil || !found {
			return false
		}
		return true
	case resourcepb.Operator_DoesNotExist:
		_, found, err := cel.GetValue(data, r.Expression)
		if err != nil || !found {
			return true
		}
		return false
	case resourcepb.Operator_GreaterThan, resourcepb.Operator_LessThan:
		v, found, err := cel.GetValue(data, r.Expression)
		if err != nil || !found {
			return false
		}
		lsValue, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Error("ParseInt failed", "value", v, "expr", r.Expression)
			return false
		}

		// There should be only one strValue in r.strValues, and can be converted to an integer.
		if len(r.Values) != 1 {
			log.Error("invalid requirement: for 'Gt', 'Lt' operators, exactly one value is required", "values", r.Values)
			return false
		}

		var rValue int64
		for i := range r.Values {
			rValue, err = strconv.ParseInt(r.Values[i], 10, 64)
			if err != nil {
				log.Error("parseInt failed for 'Gt', 'Lt' operators, the value must be an integer", "values", r.Values)
				return false
			}
		}
		return (r.Operator == resourcepb.Operator_GreaterThan && lsValue > rValue) || (r.Operator == resourcepb.Operator_LessThan && lsValue < rValue)
	default:
		return false
	}
}

func (r *Requirement) hasValue(value string) bool {
	for i := range r.Values {
		if r.Values[i] == value {
			return true
		}
	}
	return false
}

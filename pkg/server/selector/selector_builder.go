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
	"fmt"
	"strings"

	"github.com/kform-dev/choreo/pkg/server/selector/cel"
)

type SelectorBuilder interface {
	// Matches returns true if the expression match both objectss.
	GetSelector(map[string]any) map[string]string

	// Add adds requirements to the Selector
	Add(r ...BuilderRequirement) SelectorBuilder

	String() string
}

// NewParentChildSelector returns a nil parent child selector
func NewSelectorBuilder() SelectorBuilder {
	return internalSelectorBuilder(nil)
}

type internalSelectorBuilder BuilderRequirements

// String returns a comma-separated string of all
// the internalSelector Requirements' human-readable strings.
func (s internalSelectorBuilder) String() string {
	var reqs []string
	for ix := range s {
		reqs = append(reqs, s[ix].String())
	}
	return strings.Join(reqs, ",")
}

// Add adds requirements to the selector. It copies the current selector returning a new one
func (s internalSelectorBuilder) Add(reqs ...BuilderRequirement) SelectorBuilder {
	ret := make(internalSelectorBuilder, 0, len(s)+len(reqs))
	ret = append(ret, s...)
	ret = append(ret, reqs...)
	//sort.Slice(func())
	return ret
}

func (r internalSelectorBuilder) GetSelector(parent map[string]any) map[string]string {
	matches := make(map[string]string, len(r))
	for i := range r {
		k, v, err := r[i].GetMatch(parent)
		if err != nil {
			return nil
		}
		matches[k] = v
	}
	return matches
}

// Matches returns true if the Requirement matches the input Labels.
func (r *BuilderRequirement) GetMatch(parent map[string]any) (string, string, error) {
	childValue, found, err := cel.GetValue(parent, r.Expression)
	if err != nil || !found {
		return "", "", fmt.Errorf("not found")
	}
	if len(r.Values) != 1 {
		return "", "", fmt.Errorf("invalid values") // was already validated
	}
	return r.Values[0], childValue, nil

}

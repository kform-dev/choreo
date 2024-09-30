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
	"testing"

	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/util/object"
)

var input = `
apiVersion: v1
kind: Dummy
metadata:
  name: example
  namespace: network-system
  data: 1
  labels:
    example.test/dummy: a
spec:
  list:
  - name: a
    value: a1
  - name: b
    value: b1
status: 
  conditions: 
  - type: IPClaimReady
    status: "True"
  - type: BGPReady
    status: "False"
`

func TestMatches(t *testing.T) {
	cases := map[string]struct {
		selector *resourcepb.ExpressionSelector
		match    bool
	}{
		/*
			"Nil": {
				selector: nil,
				match:    false,
			},
			"Empty": {
				selector: &resourcepb.ExpressionSelector{},
				match:    true,
			},
			"Match": {
				selector: &resourcepb.ExpressionSelector{
					Match: map[string]string{
						"kind":                                  "Dummy",
						"metadata.name":                         "example",
						"metadata.labels['example.test/dummy']": "a",
					},
				},
				match: true,
			},
			"NoMatch": {
				selector: &resourcepb.ExpressionSelector{
					Match: map[string]string{
						"kind":                                  "Dummy",
						"metadata.noname":                       "example1",
						"metadata.labels['example.test/dummy']": "a",
					},
				},
				match: false,
			},
			"Expressions": {
				selector: &resourcepb.ExpressionSelector{
					MatchExpressions: []*resourcepb.ExpressionSelectorRequirement{
						{
							Expression: "kind",
							Operator:   resourcepb.Operator_In,
							Values:     []string{"Dummy"},
						},
						{
							Expression: "metadata.data",
							Operator:   resourcepb.Operator_GreaterThan,
							Values:     []string{"0"},
						},
					},
				},
				match: true,
			},
		*/
		"StatusMatch": {
			selector: &resourcepb.ExpressionSelector{
				Match: map[string]string{
					"status.conditions.exists(c, c.type == 'IPClaimReady' && c.status == 'True')": "true",
				},
			},
			match: true,
		},
		"StatusNoMatch": {
			selector: &resourcepb.ExpressionSelector{
				Match: map[string]string{
					"status.conditions.exists(c, c.type == 'BGPReady' && c.status == 'True')": "true",
				},
			},
			match: false,
		},
		"StatusNoMatchBadField": {
			selector: &resourcepb.ExpressionSelector{
				Match: map[string]string{
					"status2.conditions.exists(c, c.type == 'IPClaimReady' && c.status == 'True')": "true",
				},
			},
			match: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			u, err := object.GetUnstructuredContent([]byte(input), object.ContentTypeYAML)
			if err != nil {
				t.Errorf("yaml parse error: %s", err)
				return
			}

			selector, err := ResourceExprSelectorAsSelector(tc.selector)
			if err != nil {
				t.Errorf("selector parsing failed: %s", err)
				return
			}

			match := selector.Matches(u.Object)
			if match != tc.match {
				t.Errorf("expected match %t, got %t", tc.match, match)
			}
		})
	}
}

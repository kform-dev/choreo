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
	"testing"

	selector1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	"github.com/kform-dev/choreo/pkg/util/object"
)

var child = `
apiVersion: example.com/v1alpha1
kind: EndPoint
metadata:
  name: ixrd3.srlinux.nokia.com
  namespace: default
  ownerReferences:
  - apiVersion: dummmy.com/v1alpha1
    kind: Node
    uid: 1b48dcd9-7de8-4f10-89df-16353de51762
    name: node1
spec:
  provider: srlinux.nokia.com
  nodeType: ixrd3
status: 
  conditions: 
  - type: Ready1
    status: "True"
  - type: Ready2
    status: "False"
`

func TestParentChildMatches(t *testing.T) {
	cases := map[string]struct {
		selector *selector1alpha1.ExpressionSelector
		match    bool
	}{
		/*
			"Match": {
				selector: &choreov1alpha1.ExpressionSelector{
					Match: map[string]string{
						"spec.provider": "spec.provider",
						"spec.nodeType": "spec.nodeType",
					},
				},
				match: true,
			},
			"MatchOwnerReferences": {
				selector: &choreov1alpha1.ExpressionSelector{
					Match: map[string]string{
						"metadata.ownerReferences.filter(x, x.apiVersion == 'dummmy.com/v1alpha1')[0].uid": "metadata.uid",
					},
				},
				match: true,
			},
		*/
		"MatchStatus": {
			selector: &selector1alpha1.ExpressionSelector{
				Match: map[string]string{
					"status.conditions.exists(c, c.type == 'Ready1' && c.status == 'True')": "true",
				},
			},
			match: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			child, err := object.GetUnstructuredContent([]byte(child), object.ContentTypeYAML)
			if err != nil {
				t.Errorf("yaml parse error: %s", err)
				return
			}

			selector, err := SelectorAsSelectorBuilder(tc.selector)
			if err != nil {
				t.Errorf("selector parsing failed: %s", err)
				return
			}

			matches := selector.GetSelector(child.Object)
			fmt.Println("listSelecor", matches)
		})
	}
}

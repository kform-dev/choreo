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

package cel

import (
	"fmt"
	"testing"

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

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		input       string
		expression  string
		expectedErr bool
	}{
		/*
			"MapKeyOK": {
				input:       input,
				expression:  "metadata.labels['example.test/dummy']",
				expectedErr: false,
			},
			"MapKeyNOK": {
				input:       input,
				expression:  "metadata.labels['example.com/dummy']",
				expectedErr: true,
			},
			"StringKeyNOK": {
				input:       input,
				expression:  "metadata.key",
				expectedErr: true,
			},
			"ListWithKeyOK": {
				input:       input,
				expression:  "spec.list.filter(x, x.name == 'b')[0].value",
				expectedErr: false,
			},
			"ListWithKeyNOK": {
				input:       input,
				expression:  "spec.list.filter(x, x.name == 'c')[0].value",
				expectedErr: true,
			},
			"IntKeyNOK": {
				input:       input,
				expression:  "metadata.data",
				expectedErr: true,
			},
		*/
		"Status": {
			input:       input,
			expression:  "status.conditions.exists(c, c.type == 'IPClaimReady' && c.status == 'True')",
			expectedErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			u, err := object.GetUnstructuredContent([]byte(tc.input), object.ContentTypeYAML)
			if err != nil {
				t.Errorf("yaml parse error: %s", err)
				return
			}
			v, found, err := GetValue(u.Object, tc.expression)
			if err != nil {
				t.Errorf("unexpected error: %s", err.Error())
				return
			}
			if !found {
				if tc.expectedErr {
					return
				}
				t.Errorf("expected value but got none for expression %s", tc.expression)
				return
			}
			fmt.Printf("value %s\n", v)
		})
	}
}

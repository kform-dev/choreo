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

package genericclioptions

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

func TestToProxy(t *testing.T) {
	cases := map[string]struct {
		proxy *string
		nsn   types.NamespacedName
	}{
		"Nil": {
			proxy: nil,
			nsn:   types.NamespacedName{},
		},
		"Empty": {
			proxy: ptr.To(""),
			nsn:   types.NamespacedName{},
		},
		"Single": {
			proxy: ptr.To("a"),
			nsn:   types.NamespacedName{Name: "a", Namespace: "default"},
		},
		"Two": {
			proxy: ptr.To("a.b"),
			nsn:   types.NamespacedName{Name: "b", Namespace: "a"},
		},
		"Multiple": {
			proxy: ptr.To("a.b.c.d"),
			nsn:   types.NamespacedName{Name: "b.c.d", Namespace: "a"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			flags := &ConfigFlags{
				Proxy: tc.proxy,
			}
			nsn := flags.ToProxy()
			if cmp.Diff(nsn, tc.nsn) != "" {
				t.Errorf("%s, expected %s got %s", name, tc.nsn.String(), nsn.String())
			}
		})
	}
}

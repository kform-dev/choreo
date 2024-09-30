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

package starlark

import (
	"fmt"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func TestCache(t *testing.T) {
	cases := map[string]struct {
		cache       *cache
		reconciler  string
		expextedErr bool
	}{
		"NonCycle": {
			cache: &cache{
				cache: make(map[string]*entry),
				modules: map[string]string{
					"a.star": `load("b.star", "b"); a = b + 1`,
					"b.star": `b=10`,
				},
			},
			reconciler:  `load("a.star", "a"); print(a)`,
			expextedErr: false,
		},
		"Cycle": {
			cache: &cache{
				cache: make(map[string]*entry),
				modules: map[string]string{
					"a.star": `load("b.star", "b"); a = b + 1`,
					"b.star": `load("a.star", "a"); b = a + 1`,
				},
			},
			reconciler:  `load("a.star", "a"); print(a)`,
			expextedErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cc := new(cycleChecker)
			builtins := starlark.StringDict{}
			thread := &starlark.Thread{
				Name: "main",
				Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
					return tc.cache.get(cc, module, builtins)
				},
			}
			if _, err := starlark.ExecFileOptions(&syntax.FileOptions{}, thread, "reconciler.star", tc.reconciler, builtins); err != nil {
				if !tc.expextedErr {
					t.Errorf("unexpected cyclic error %s", err.Error())
				}
				fmt.Println(err)
				return
			}
			if tc.expextedErr {
				t.Errorf("expected cyclic error got nil")
			}

		})
	}
}

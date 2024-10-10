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

package parser

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestRender(t *testing.T) {
	cases := map[string]struct {
		pathTemplates       string
		pathToInputYAMLFile string
	}{
		"Normal": {
			pathTemplates:       "./test1/templates",
			pathToInputYAMLFile: "./test1/data/interface.yaml",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			reader := loader.GetFSReader(tc.pathTemplates)
			templateStorer, err := reader.Read(ctx)
			if err != nil {
				t.Errorf("reader failed: %v", err)
				return
			}

			templateFiles := map[string]string{}
			templateStorer.List(func(k store.Key, b []byte) {
				templateFiles[k.Name] = string(b)
			})

			b, err := os.ReadFile(tc.pathToInputYAMLFile)
			if err != nil {
				t.Errorf("cannot read input yaml file %s %v", tc.pathToInputYAMLFile, err)
				return
			}
			data := map[string]any{}
			if err := yaml.Unmarshal(b, &data); err != nil {
				t.Errorf("cannot unmarshal yaml file %s %v", tc.pathToInputYAMLFile, err)
				return
			}

			u := &unstructured.Unstructured{
				Object: data,
			}

			fmt.Println("input\n", u)

			fmt.Println("templateFiles", templateFiles)

			p, err := New(templateFiles)
			if err != nil {
				t.Errorf("creating parser failed: %v", err)
				return
			}

			var buf bytes.Buffer
			if err := p.Render(ctx, "main.tpl", u, &buf); err != nil {
				t.Errorf("failed rendering data: %v", err)
				return
			}

			fmt.Println(buf.String())

			data = map[string]any{}
			if err := yaml.Unmarshal(buf.Bytes(), &data); err != nil {
				t.Errorf("failed rendering data: %v", err)
				return
			}
		})
	}
}

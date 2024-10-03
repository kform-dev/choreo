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
	"context"
	"fmt"
	"io"
	"os"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func New(files map[string]string) (*Parser, error) {
	tmpl := template.New("main").Funcs(templateHelperFunctions)
	for filename, data := range files {
		_, err := tmpl.New(filename).Parse(data)
		if err != nil {
			return nil, fmt.Errorf("cannot parse template %s", err.Error())
		}
	}
	return &Parser{
		tmpl: tmpl,
	}, nil
}

type Parser struct {
	tmpl *template.Template
}

func (r *Parser) Render(ctx context.Context, template string, u *unstructured.Unstructured, w io.Writer) error {
	if w != nil {
		return r.tmpl.ExecuteTemplate(w, template, u.Object)
	}
	return r.tmpl.ExecuteTemplate(os.Stdout, template, u.Object)
}

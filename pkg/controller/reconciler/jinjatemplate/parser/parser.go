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
	"io"
	"os"

	"github.com/flosch/pongo2/v6"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Loader struct {
	templates map[string]string
}

func (r *Loader) Abs(base, name string) string {
	// returns the name of the file
	return name
}

func (r *Loader) Get(name string) (io.Reader, error) {
	// return the content of the file
	buf := bytes.NewBuffer([]byte(r.templates[name]))
	return buf, nil

}

func New(templates map[string]string) (*Parser, error) {
	ts := pongo2.NewSet("choreo", &Loader{
		templates: templates,
	})

	mainTemplate, err := ts.FromString(templates["main.jinja2"])
	if err != nil {
		return nil, err
	}

	return &Parser{
		tmpl: mainTemplate,
	}, nil
}

type Parser struct {
	tmpl *pongo2.Template
}

func (r *Parser) Render(ctx context.Context, template string, u *unstructured.Unstructured, w io.Writer) error {
	if w != nil {
		return r.tmpl.ExecuteWriter(u.Object, w)
	}
	return r.tmpl.ExecuteWriter(u.Object, os.Stdout)
}

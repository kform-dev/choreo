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
		return r.tmpl.ExecuteTemplate(w, template, u.Object["spec"])
	}
	return r.tmpl.ExecuteTemplate(os.Stdout, template, u.Object["spec"])
}

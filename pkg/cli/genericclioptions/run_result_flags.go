/*
Copyright 2018 The Kubernetes Authors.

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
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
)

const (
	FlagResultOutputFormat = "result-output-format"
)

var (
	SupportedResultOutputFormats = sets.New[string]().Insert([]string{"reconciler", "resource", "raw"}...)
)

// ResourceFlags are flags for generic resources.
type RunResultFlags struct {
	ResultOutputFormat *string
}

func NewRunResultFlags() *RunResultFlags {
	return &RunResultFlags{
		ResultOutputFormat: ptr.To("reconciler"),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *RunResultFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}

	if r.ResultOutputFormat != nil {
		flags.StringVarP(r.ResultOutputFormat, FlagShowChoreoAPIs, "o", *r.ResultOutputFormat,
			"result output is either 'reconciler', 'resource' or 'raw")
	}
}

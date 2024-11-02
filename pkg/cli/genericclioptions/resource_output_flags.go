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
	"k8s.io/utils/ptr"
)

const (
	FlagOutputFormat      = "output"
	FlagShowManagedFields = "show-managed-fields"
)

// ResourceFlags are flags for generic resources.
type ResourceOutputFlags struct {
	Output            *string
	ShowManagedFields *bool
}

func NewResourceOutputFlags() *ResourceOutputFlags {
	return &ResourceOutputFlags{
		Output:            ptr.To(""),
		ShowManagedFields: ptr.To(false),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *ResourceOutputFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}

	if r.Output != nil {
		flags.StringVarP(r.Output, FlagOutputFormat, "o", *r.Output,
			"output is either 'json' or 'yaml'")
	}
	if r.ShowManagedFields != nil {
		flags.BoolVar(r.ShowManagedFields, FlagShowManagedFields, *r.ShowManagedFields,
			"if true, keep the managedFields when printing objects in JSON or YAML format.")
	}
}

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
	FlagShowChoreoAPIs  = "show-choreo-apis"
	FlagShowDiffDetails = "show-diff-details"
)

// ResourceFlags are flags for generic resources.
type RunOutputFlags struct {
	ShowChoreoAPIs    *bool
	ShowManagedFields *bool
	ShowDiffDetails   *bool
}

func NewRunOutputFlags() *RunOutputFlags {
	return &RunOutputFlags{
		ShowChoreoAPIs:    ptr.To(false),
		ShowManagedFields: ptr.To(false),
		ShowDiffDetails:   ptr.To(false),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *RunOutputFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}

	if r.ShowChoreoAPIs != nil {
		flags.BoolVar(r.ShowChoreoAPIs, FlagShowChoreoAPIs, *r.ShowChoreoAPIs,
			"if true, returns besides the custom apis also the internal choreo apis")
	}
	if r.ShowManagedFields != nil {
		flags.BoolVar(r.ShowManagedFields, FlagShowManagedFields, *r.ShowManagedFields,
			"if true, keep the managedFields when printing objects in JSON or YAML format.")
	}
	if r.ShowDiffDetails != nil {
		flags.BoolVarP(r.ShowDiffDetails, FlagShowDiffDetails, "a", *r.ShowDiffDetails,
			"if true, show the detailed diff when the resources between the snapshots are different")
	}
}

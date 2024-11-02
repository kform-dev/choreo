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
	FlagNamespace    = "namespace"
	defaultNamespace = "default"
)

// ResourceFlags are flags for generic resources.
type ResourceFlags struct {
	Namespace *string
}

func NewResourceFlags() *ResourceFlags {
	return &ResourceFlags{
		Namespace: ptr.To("default"),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *ResourceFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}

	if r.Namespace != nil {
		flags.StringVarP(r.Namespace, FlagNamespace, "n", *r.Namespace, "If present, the namespace scope for this CLI request")
	}
}

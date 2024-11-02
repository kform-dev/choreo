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

package tuicmd

import (
	"github.com/spf13/pflag"
	"k8s.io/utils/ptr"
)

const (
	flagFrequency = "frequency"
)

// ResourceFlags are flags for generic resources.
type tuiFlags struct {
	frequency *float64
}

func newtuiflags() *tuiFlags {
	return &tuiFlags{
		frequency: ptr.To(float64(3.0)),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *tuiFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}

	if r.frequency != nil {
		flags.Float64VarP(r.frequency, flagFrequency, "f", *r.frequency, "refresh frequency in seconds")
	}
}

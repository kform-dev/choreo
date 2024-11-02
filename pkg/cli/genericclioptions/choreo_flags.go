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

package genericclioptions

import (
	"path/filepath"

	"github.com/spf13/pflag"
	"k8s.io/utils/ptr"
)

const (
	flagDebug   = "debug"
	flagAddress = "address"
	flagConfig  = "config"
)

const (
	defaultConfigFileSubDir = "choreoctl"
	defaultConfigFileName   = "config.yaml"
	defaultConfigEnvPrefix  = "CHOREOCTL"
)

type ChoreoFlags struct {
	Debug           *bool
	Address         *string
	Config          *string
	ConfigEnvPrefix *string
}

// NewConfigFlags returns ConfigFlags with default values set
func NewChoreoFlags() *ChoreoFlags {
	return &ChoreoFlags{
		// relevant for all
		Debug:   ptr.To(false),
		Address: ptr.To("0.0.0.0:51000"),
		Config:  ptr.To(filepath.Join(getConfigPath(), defaultConfigFileName)),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *ChoreoFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}

	if r.Debug != nil {
		flags.BoolVarP(r.Debug, flagDebug, "d", *r.Debug,
			"enable debug mode")
	}
	if r.Address != nil {
		flags.StringVar(r.Address, flagAddress, *r.Address,
			"address the server is listing on")
	}
	if r.Config != nil {
		flags.StringVar(r.Config, flagConfig, *r.Config,
			"configuration where the client context is stored")
	}
}

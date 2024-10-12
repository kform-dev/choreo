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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func (r *ConfigFlags) AddClientFlags(fs *pflag.FlagSet) {
	if r == nil {
		return
	}
	if r.Debug != nil {
		fs.BoolVarP(r.Debug, flagDebug, "d", *r.Debug,
			"enable debug mode")
	}
	if r.Address != nil {
		fs.StringVar(r.Address, flagAddress, *r.Address,
			"address the server is listing on")
	}
	if r.Output != nil {
		fs.StringVarP(r.Output, FlagOutputFormat, "o", *r.Output,
			"output is either 'json' or 'yaml'")
	}
	if r.MaxRcvMsg != nil {
		fs.IntVar(r.MaxRcvMsg, flagMaxRcvMsg, *r.MaxRcvMsg,
			"the maximum message size in bytes the client can receive")
	}
	if r.Timeout != nil {
		fs.IntVar(r.Timeout, flagTimeout, *r.Timeout,
			"gRPC dial timeout in seconds")
	}
	if r.Config != nil {
		fs.StringVar(r.Config, flagConfig, *r.Config,
			"configuration where the client context is stored")
	}
	if r.CacheDir != nil {
		fs.StringVar(r.CacheDir, flagCacheDir, *r.CacheDir,
			"cache directory where the api resource information is stored")
	}
	if r.Branch != nil {
		fs.StringVarP(r.Branch, flagBranch, "b", *r.Branch,
			"branch from which the client wants to retrieve the info")
	}
	if r.Proxy != nil {
		fs.StringVarP(r.Proxy, flagProxy, "p", *r.Proxy,
			"proxy context from which the client wants to retrieve the info")
	}
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig(cmd *cobra.Command) error {
	config, err := cmd.Flags().GetString(flagConfig)
	if err != nil {
		return err
	}

	path := filepath.Dir(config)
	base := filepath.Base(config)
	ext := filepath.Ext(config)
	filename := base[:len(base)-len(ext)]

	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}

	viper.AddConfigPath(path)
	viper.SetConfigType(ext)
	viper.SetConfigName(filename)

	//err := viper.SafeWriteConfig()

	//viper.Set("kubecontext", kubecontext)
	//viper.Set("kubeconfig", kubeconfig)

	viper.SetEnvPrefix(defaultConfigEnvPrefix)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_ = 1
	}
	return nil
}

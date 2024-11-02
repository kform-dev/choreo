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

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

func getConfigPath() string {
	return filepath.Join(xdg.ConfigHome, defaultConfigFileSubDir)
}

// getDefaultCacheDir returns default caching directory path.
// it first looks at KUBECACHEDIR env var if it is set, otherwise
// it returns standard kube cache dir.
func getDefaultCacheDir(configPath string) string {
	return filepath.Join(configPath, "cache")
}

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
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/client/go/discovery/cached/disk"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/client/go/resourcemapper"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
	"github.com/kform-dev/choreo/pkg/client/go/snapshotclient"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/types"
)

var _ ClientGetter = &ChoreoConfig{}

type ChoreoConfig struct {
	ChoreoFlags *ChoreoFlags
	ClientFlags *ClientFlags
	ServerFlags *ServerFlags
}

func NewChoreoConfig() *ChoreoConfig {
	return &ChoreoConfig{
		ChoreoFlags: NewChoreoFlags(),
		ClientFlags: NewClientFlags(),
		ServerFlags: NewServerFlags(),
	}
}

func (r *ChoreoConfig) AddFlags(flags *pflag.FlagSet) {
	r.ChoreoFlags.AddFlags(flags)
	r.ClientFlags.AddFlags(flags)
	r.ServerFlags.AddFlags(flags)
}

func (r *ChoreoConfig) ToConfig() *config.Config {
	return r.toConfig()
}

func (r *ChoreoConfig) toConfig() *config.Config {
	return &config.Config{
		Address:    *r.ChoreoFlags.Address,
		MaxMsgSize: *r.ClientFlags.MaxRcvMsg,
		Timeout:    time.Duration(*r.ClientFlags.Timeout) * time.Second,
	}
}

func (r *ChoreoConfig) ToChoreoClient() (choreoclient.Client, error) {
	return r.toChoreoClient()
}

func (r *ChoreoConfig) toChoreoClient() (choreoclient.Client, error) {
	config := r.toConfig()
	return choreoclient.NewClient(config)
}

func (r *ChoreoConfig) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return r.toDiscoveryClient()
}

func (r *ChoreoConfig) toDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config := r.toConfig()

	parts := strings.Split(*r.ChoreoFlags.Address, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid address, %s", *r.ChoreoFlags.Address)
	}

	discoveryCacheDir := filepath.Join(*r.ClientFlags.CacheDir, "discovery", parts[0], parts[1])

	return disk.NewCachedDiscoveryClient(config, discoveryCacheDir, time.Duration(6*time.Hour))
}

func (r *ChoreoConfig) ToResourceClient() (resourceclient.Client, error) {
	return r.toResourceClient()
}

func (r *ChoreoConfig) toResourceClient() (resourceclient.Client, error) {
	config := r.toConfig()
	return resourceclient.NewClient(config)
}

func (r *ChoreoConfig) ToResourceMapper() (resourcemapper.Mapper, error) {
	return r.toResourceMapper()
}

func (r *ChoreoConfig) toResourceMapper() (resourcemapper.Mapper, error) {
	discoveryClient, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	return resourcemapper.NewMapper(discoveryClient), nil
}

func (r *ChoreoConfig) ToBranchClient() (branchclient.Client, error) {
	config := r.toConfig()
	return branchclient.NewClient(config)
}

func (r *ChoreoConfig) ToRunnerClient() (runnerclient.Client, error) {
	config := r.toConfig()
	return runnerclient.NewClient(config)
}

func (r *ChoreoConfig) ToSnapshotClient() (snapshotclient.Client, error) {
	config := r.toConfig()
	return snapshotclient.NewClient(config)
}

func (r *ChoreoConfig) ToBranch() string {
	if r.ClientFlags.Branch == nil {
		return ""
	}
	return *r.ClientFlags.Branch
}

func (r *ChoreoConfig) ToProxy() types.NamespacedName {
	if r.ClientFlags.Proxy == nil {
		return types.NamespacedName{}
	}
	if *r.ClientFlags.Proxy == "" {
		return types.NamespacedName{}
	}
	parts := strings.SplitN(*r.ClientFlags.Proxy, ".", 2)
	if len(parts) == 1 {
		return types.NamespacedName{
			Name:      *r.ClientFlags.Proxy,
			Namespace: defaultNamespace,
		}
	}

	return types.NamespacedName{
		Name:      parts[1],
		Namespace: parts[0],
	}

}

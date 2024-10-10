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

	"github.com/adrg/xdg"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/client/go/discovery/cached/disk"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/client/go/resourcemapper"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
	"k8s.io/utils/ptr"
)

const (
	flagDebug               = "debug"
	flagAddress             = "address"
	flagServerName          = "servername"
	flagAPIs                = "apis"
	flagDB                  = "db"
	flagReconcilers         = "reconcilers"
	flagLibraries           = "libraries"
	flagInput               = "input"
	flagPostprocessing      = "post"
	flagOutput              = "output"
	flagRefs                = "refs"
	flagMaxRcvMsg           = "maxRcvMsg"
	flagTimeout             = "timeout"
	FlagOutputFormat        = "output"
	flagConfig              = "config"
	flagCacheDir            = "cacheDir"
	flagInternalReconcilers = "internalReconcilers"
	flagBranch              = "branch"
)

const (
	defaultConfigFileSubDir = "choreoctl"
	defaultConfigFileName   = "config.yaml"
	defaultConfigEnvPrefix  = "CHOREOCTL"
)

type ClientGetter interface {
	// ToConfig returns config
	ToConfig() *config.Config
	// ToDiscoveryClient returns discovery client
	ToChoreoClient() (choreoclient.Client, error)
	// ToDiscoveryClient returns discovery client
	ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error)
	// ToResourceMapper returns a restmapper
	ToResourceMapper() (resourcemapper.Mapper, error)
	// ToResourceClient returns resource client
	ToResourceClient() (resourceclient.Client, error)
	// ToBranchClient returns branch client
	ToBranchClient() (branchclient.Client, error)
	// ToRunnerClient returns runner client
	ToRunnerClient() (runnerclient.Client, error)
	// Branch()
	ToBranch() string
}

var _ ClientGetter = &ConfigFlags{}

type ConfigFlags struct {
	Debug               *bool
	Address             *string
	CRDPath             *string
	DBPath              *string
	ReconcilerPath      *string
	LibraryPath         *string
	PostProcessingPath  *string
	OutputPath          *string
	InputPath           *string
	RefsPath            *string
	MaxRcvMsg           *int
	Timeout             *int
	CacheDir            *string
	Output              *string
	Config              *string
	ConfigEnvPrefix     *string
	InternalReconcilers *bool
	Branch              *string
}

// NewConfigFlags returns ConfigFlags with default values set
func NewConfigFlags() *ConfigFlags {
	configPath := filepath.Join(xdg.ConfigHome, defaultConfigFileSubDir)
	return &ConfigFlags{
		Debug:               ptr.To(false),
		Address:             ptr.To("0.0.0.0:51000"),
		CRDPath:             ptr.To("crds"),
		DBPath:              ptr.To("db"),
		ReconcilerPath:      ptr.To("reconcilers"),
		LibraryPath:         ptr.To("libs"),
		PostProcessingPath:  ptr.To("post"),
		OutputPath:          ptr.To("out"),
		InputPath:           ptr.To("in"),
		RefsPath:            ptr.To("refs"),
		MaxRcvMsg:           ptr.To(25165824),
		Timeout:             ptr.To(10),
		CacheDir:            ptr.To(getDefaultCacheDir(configPath)),
		Output:              ptr.To(""),
		Config:              ptr.To(filepath.Join(configPath, defaultConfigFileName)),
		InternalReconcilers: ptr.To(false),
		Branch:              ptr.To(""),
	}
}

// getDefaultCacheDir returns default caching directory path.
// it first looks at KUBECACHEDIR env var if it is set, otherwise
// it returns standard kube cache dir.
func getDefaultCacheDir(configPath string) string {
	return filepath.Join(configPath, "cache")
}

func (r *ConfigFlags) ToConfig() *config.Config {
	return r.toConfig()
}

func (r *ConfigFlags) toConfig() *config.Config {
	return &config.Config{
		Address:    *r.Address,
		MaxMsgSize: *r.MaxRcvMsg,
		Timeout:    time.Duration(*r.Timeout) * time.Second,
	}
}

func (r *ConfigFlags) ToChoreoClient() (choreoclient.Client, error) {
	return r.toChoreoClient()
}

func (r *ConfigFlags) toChoreoClient() (choreoclient.Client, error) {
	config := r.toConfig()
	return choreoclient.NewClient(config)
}

func (r *ConfigFlags) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return r.toDiscoveryClient()
}

func (r *ConfigFlags) toDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config := r.toConfig()

	parts := strings.Split(*r.Address, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid address, %s", *r.Address)
	}

	discoveryCacheDir := filepath.Join(*r.CacheDir, "discovery", parts[0], parts[1])

	return disk.NewCachedDiscoveryClient(config, discoveryCacheDir, time.Duration(6*time.Hour))
}

func (r *ConfigFlags) ToResourceClient() (resourceclient.Client, error) {
	return r.toResourceClient()
}

func (r *ConfigFlags) toResourceClient() (resourceclient.Client, error) {
	config := r.toConfig()
	return resourceclient.NewClient(config)
}

func (r *ConfigFlags) ToResourceMapper() (resourcemapper.Mapper, error) {
	return r.toResourceMapper()
}

func (r *ConfigFlags) toResourceMapper() (resourcemapper.Mapper, error) {
	discoveryClient, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	return resourcemapper.NewMapper(discoveryClient), nil
}

func (r *ConfigFlags) ToBranchClient() (branchclient.Client, error) {
	config := r.toConfig()
	return branchclient.NewClient(config)
}

func (r *ConfigFlags) ToBranch() string {
	return *r.Branch
}

func (r *ConfigFlags) ToRunnerClient() (runnerclient.Client, error) {
	config := r.toConfig()
	return runnerclient.NewClient(config)
}

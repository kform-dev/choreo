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
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/client/go/resourcemapper"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
	"github.com/kform-dev/choreo/pkg/client/go/snapshotclient"
	"k8s.io/apimachinery/pkg/types"
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
	// TosnapshotClient returns snapshot client
	ToSnapshotClient() (snapshotclient.Client, error)
	// Branch()
	ToBranch() string
	// Proxy()
	ToProxy() types.NamespacedName
}

type ConfigFn func() *config.Config
type ResourceClientFn func() (resourceclient.Client, error)
type ResourceMapperFn func() (resourcemapper.Mapper, error)

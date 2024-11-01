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

type NoopClientGetter struct{}

func (NoopClientGetter) ToConfig() *config.Config {
	return nil
}

// ToDiscoveryClient returns discovery client
func (NoopClientGetter) ToChoreoClient() (choreoclient.Client, error) {
	return nil, fmt.Errorf("local operation only")
}

// ToDiscoveryClient returns discovery client
func (NoopClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return nil, fmt.Errorf("local operation only")
}

// ToResourceMapper returns a restmapper
func (NoopClientGetter) ToResourceMapper() (resourcemapper.Mapper, error) {
	return nil, fmt.Errorf("local operation only")
}

// ToResourceClient returns resource client
func (NoopClientGetter) ToResourceClient() (resourceclient.Client, error) {
	return nil, fmt.Errorf("local operation only")
}

// ToBranchClient returns branch client
func (NoopClientGetter) ToBranchClient() (branchclient.Client, error) {
	return nil, fmt.Errorf("local operation only")
}

// ToRunnerClient returns runner client
func (NoopClientGetter) ToRunnerClient() (runnerclient.Client, error) {
	return nil, fmt.Errorf("local operation only")
}

// TosnapshotClient returns snapshot client
func (NoopClientGetter) ToSnapshotClient() (snapshotclient.Client, error) {
	return nil, fmt.Errorf("local operation only")
}

// Branch()
func (NoopClientGetter) ToBranch() string { return "" }

// Proxy()
func (NoopClientGetter) ToProxy() types.NamespacedName { return types.NamespacedName{} }

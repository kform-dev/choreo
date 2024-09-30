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

package util

import (
	"errors"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/client/go/resourcemapper"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
)

type Factory interface {
	GetConfig() *config.Config
	GetDiscoveryClient() discovery.CachedDiscoveryInterface
	GetResourceMapper() resourcemapper.Mapper
	GetResourceClient() resourceclient.Client
	GetBranchClient() branchclient.Client
	GetRunnerClient() runnerclient.Client
	Close() error
}

func NewFactory(clientGetter genericclioptions.ClientGetter) (Factory, error) {
	if clientGetter == nil {
		panic("attempt to instantiate factory with nil clientGetter")
	}

	config := clientGetter.ToConfig()

	discoveryCLient, err := clientGetter.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	branchClient, err := clientGetter.ToBranchClient()
	if err != nil {
		return nil, err
	}

	resourceClient, err := clientGetter.ToResourceClient()
	if err != nil {
		return nil, err
	}

	runnerClient, err := clientGetter.ToRunnerClient()
	if err != nil {
		return nil, err
	}

	resourceMapper := resourcemapper.NewMapper(discoveryCLient)

	return &factory{
		config:          config,
		discoveryCLient: discoveryCLient,
		resourceMapper:  resourceMapper,
		resourceClient:  resourceClient,
		branchClient:    branchClient,
		runnerClient:    runnerClient,
	}, nil
}

type factory struct {
	config          *config.Config
	discoveryCLient discovery.CachedDiscoveryInterface
	resourceMapper  resourcemapper.Mapper
	resourceClient  resourceclient.Client
	branchClient    branchclient.Client
	runnerClient    runnerclient.Client
}

func (r *factory) Close() error {
	var errm error
	if err := r.discoveryCLient.Close(); err != nil {
		errm = errors.Join(errm, err)
	}

	if err := r.resourceClient.Close(); err != nil {
		errm = errors.Join(errm, err)
	}
	return errm
}

func (r *factory) GetConfig() *config.Config {
	return r.config
}

func (r *factory) GetDiscoveryClient() discovery.CachedDiscoveryInterface {
	return r.discoveryCLient
}

func (r *factory) GetResourceMapper() resourcemapper.Mapper {
	return r.resourceMapper
}

func (r *factory) GetResourceClient() resourceclient.Client {
	return r.resourceClient
}

func (r *factory) GetBranchClient() branchclient.Client {
	return r.branchClient
}

func (r *factory) GetRunnerClient() runnerclient.Client {
	return r.runnerClient
}

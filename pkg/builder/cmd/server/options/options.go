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

package options

import (
	"context"
	"errors"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"github.com/kform-dev/choreo/pkg/server/grpcserver"
	"github.com/kform-dev/choreo/pkg/server/health"
)

// ChoreoServer is the default name of the choreo server
const ChoreoServer = "choreoServer"

type ChoreoOptions struct {
	serverName string
	path       string
	flags      *genericclioptions.ConfigFlags
}

func NewOptions(path string, flags *genericclioptions.ConfigFlags) *ChoreoOptions {
	return &ChoreoOptions{
		serverName: ChoreoServer,
		path:       path,
		flags:      flags,
	}
}

func (r *ChoreoOptions) Flags() *genericclioptions.ConfigFlags {
	return r.flags
}

func (r *ChoreoOptions) Complete() error {
	return nil
}

func (r *ChoreoOptions) Validate(ctx context.Context) error {
	if r == nil {
		return nil
	}
	var errm error
	if r.serverName == "" {
		errm = errors.Join(errm, fmt.Errorf("a grpc servername must be specified"))
	}
	if r.flags.Address == nil {
		errm = errors.Join(errm, fmt.Errorf("an address must be specified"))
	}
	if r.flags.CRDPath == nil {
		errm = errors.Join(errm, fmt.Errorf("an api path must be specified"))
	}
	return errm
}

func (r *ChoreoOptions) Run(ctx context.Context) error {
	log := log.FromContext(ctx)

	choreo, err := choreo.New(ctx, r.path, r.flags)
	if err != nil {
		return err
	}
	// build grpc server which hosts the apiserver storage
	grpcserver := grpcserver.New(&grpcserver.Config{
		Name:   r.serverName,
		Flags:  r.flags,
		Choreo: choreo,
	})
	// start the server
	go func() {
		err := grpcserver.Run(ctx)
		if err != nil {
			log.Error("grpc server failed", "err", err)
		}
	}()
	if !health.IsServerReady(ctx, r.flags) {
		return fmt.Errorf("server is not ready")
	}

	//r.repo.AddResourceClient(client)
	ctx, cancel := context.WithCancel(ctx)
	go choreo.Start(ctx)
	defer cancel()

	<-ctx.Done()
	log.Debug("context concelled")
	return nil
}

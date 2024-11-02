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

package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/proto/snapshotpb"
	choreoserver "github.com/kform-dev/choreo/pkg/server/choreo"
	"github.com/kform-dev/choreo/pkg/server/grpcserver/services/branch"
	"github.com/kform-dev/choreo/pkg/server/grpcserver/services/choreo"
	"github.com/kform-dev/choreo/pkg/server/grpcserver/services/discovery"
	"github.com/kform-dev/choreo/pkg/server/grpcserver/services/resource"
	"github.com/kform-dev/choreo/pkg/server/grpcserver/services/runner"
	"github.com/kform-dev/choreo/pkg/server/grpcserver/services/snapshot"
	choreohealth "github.com/kform-dev/choreo/pkg/server/health"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Name   string
	Cfg    *genericclioptions.ChoreoConfig
	Choreo choreoserver.Choreo
}

func New(cfg *Config) *GRPCServer {
	grpcServer := func(opts []grpc.ServerOption) *grpc.Server {
		return grpc.NewServer(append(opts,
			//grpc.KeepaliveParams(keepalive.ServerParameters{}), // server does not ping the client
			grpc.MaxSendMsgSize(64<<20 /* 64MB */),
			grpc.MaxRecvMsgSize(64<<20 /* 64MB */))...)
	}

	return &GRPCServer{
		name:    cfg.Name,
		address: *cfg.Cfg.ChoreoFlags.Address,
		server:  grpcServer([]grpc.ServerOption{}),
		choreo:  cfg.Choreo,
		cfg:     cfg.Cfg,
	}
}

type GRPCServer struct {
	name    string
	address string
	cancel  context.CancelFunc
	server  *grpc.Server
	choreo  choreoserver.Choreo
	cfg     *genericclioptions.ChoreoConfig
}

func (r *GRPCServer) Stop(ctx context.Context) {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *GRPCServer) Run(ctx context.Context) error {
	log := log.FromContext(ctx).With("name", r.name, "address", r.address)
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	l, err := net.Listen("tcp", r.address)
	if err != nil {
		return errors.Wrap(err, "cannot listen")
	}
	// Register the health service
	healthCheck := health.NewServer()
	healthCheck.SetServingStatus(
		r.name, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(r.server, healthCheck)

	// Register the reflection service
	reflection.Register(r.server)

	// Register the resource service
	choreoServer := choreo.New(r.choreo)
	choreopb.RegisterChoreoServer(r.server, choreoServer)

	// Register the resource service
	resourceServer := resource.New(r.choreo)
	resourcepb.RegisterResourceServer(r.server, resourceServer)

	// Register the discovery service
	discoveryServer := discovery.New(r.choreo)
	discoverypb.RegisterDiscoveryServer(r.server, discoveryServer)

	// Register the branch service
	branchServer := branch.New(r.choreo)
	branchpb.RegisterBranchServer(r.server, branchServer)

	// Register the runner service
	runnerServer := runner.New(r.choreo)
	runnerpb.RegisterRunnerServer(r.server, runnerServer)

	// Register the runner service
	snapshortServer := snapshot.New(r.choreo)
	snapshotpb.RegisterSnapshotServer(r.server, snapshortServer)

	go func() {
		if err := r.server.Serve(l); err != nil {
			log.Error("grpc server serve", "error", err)
		}
	}()
	log.Info("server started")

	time.Sleep(1 * time.Second)
	if !choreohealth.IsServerReady(ctx, &choreohealth.Config{Address: r.address}) {
		return fmt.Errorf("server is not ready")
	}

	for range ctx.Done() {
		log.Info("server stopped...")
		r.cancel()
	}
	return nil
}

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

package runnerclient

import (
	"context"

	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RunnerClient interface {
	runnerpb.RunnerClient
	Close() error
}

func NewRunnerClient(config *config.Config) (RunnerClient, error) {
	client := &runnerclient{
		config: config,
	}

	conn, err := grpc.NewClient(config.Address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(config.MaxMsgSize)),
	)
	if err != nil {
		return nil, err
	}
	client.client = runnerpb.NewRunnerClient(conn)
	client.conn = conn
	return client, nil
}

type runnerclient struct {
	config *config.Config
	conn   *grpc.ClientConn
	client runnerpb.RunnerClient
}

func (r *runnerclient) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *runnerclient) Start(ctx context.Context, in *runnerpb.Start_Request, opts ...grpc.CallOption) (*runnerpb.Start_Response, error) {
	return r.client.Start(ctx, in, opts...)
}
func (r *runnerclient) Stop(ctx context.Context, in *runnerpb.Stop_Request, opts ...grpc.CallOption) (*runnerpb.Stop_Response, error) {
	return r.client.Stop(ctx, in, opts...)
}
func (r *runnerclient) Once(ctx context.Context, in *runnerpb.Once_Request, opts ...grpc.CallOption) (runnerpb.Runner_OnceClient, error) {
	return r.client.Once(ctx, in, opts...)
}
func (r *runnerclient) Load(ctx context.Context, in *runnerpb.Load_Request, opts ...grpc.CallOption) (*runnerpb.Load_Response, error) {
	return r.client.Load(ctx, in, opts...)
}

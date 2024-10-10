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

/*

import (
	"context"

	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	Start(ctx context.Context, branch string) error
	Stop(ctx context.Context, branch string) error
	Once(ctx context.Context, branch string) (*runnerpb.Once_Response, error)
	Load(ctx context.Context, branch string) error
}

func NewClient(config *config.Config) (Client, error) {
	client := &client{
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

type client struct {
	config *config.Config
	conn   *grpc.ClientConn
	client runnerpb.RunnerClient
}

func (r *client) Start(ctx context.Context, branch string) error {
	if _, err := r.client.Start(ctx, &runnerpb.Start_Request{
		Branch: branch,
	}); err != nil {
		return err
	}
	return nil
}

func (r *client) Stop(ctx context.Context, branch string) error {
	if _, err := r.client.Stop(ctx, &runnerpb.Stop_Request{
		Branch: branch,
	}); err != nil {
		return err
	}
	return nil
}

func (r *client) Once(ctx context.Context, branch string) (*runnerpb.Once_Response, error) {
	rsp, err := r.client.Once(ctx, &runnerpb.Once_Request{
		Branch: branch,
	})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (r *client) Load(ctx context.Context, branch string) error {
	if _, err := r.client.Load(ctx, &runnerpb.Load_Request{
		Branch: branch,
	}); err != nil {
		return err
	}
	return nil
}
*/

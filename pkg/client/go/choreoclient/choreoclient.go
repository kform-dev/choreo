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

package choreoclient

import (
	"context"

	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChoreoClient interface {
	choreopb.ChoreoClient
	Close() error
}

func NewChoreoClient(config *config.Config) (ChoreoClient, error) {
	client := &choreoclient{
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
	client.client = choreopb.NewChoreoClient(conn)
	client.conn = conn
	return client, nil
}

type choreoclient struct {
	config *config.Config
	conn   *grpc.ClientConn
	client choreopb.ChoreoClient
}

func (r *choreoclient) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *choreoclient) Get(ctx context.Context, in *choreopb.Get_Request, opts ...grpc.CallOption) (*choreopb.Get_Response, error) {
	return r.client.Get(ctx, in, opts...)
}
func (r *choreoclient) Apply(ctx context.Context, in *choreopb.Apply_Request, opts ...grpc.CallOption) (*choreopb.Apply_Response, error) {
	return r.client.Apply(ctx, in, opts...)
}
func (r *choreoclient) Commit(ctx context.Context, in *choreopb.Commit_Request, opts ...grpc.CallOption) (*choreopb.Commit_Response, error) {
	return r.client.Commit(ctx, in, opts...)
}
func (r *choreoclient) Push(ctx context.Context, in *choreopb.Push_Request, opts ...grpc.CallOption) (*choreopb.Push_Response, error) {
	return r.client.Push(ctx, in, opts...)
}

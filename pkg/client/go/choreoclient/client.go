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

type Client interface {
	Get(ctx context.Context, opts ...GetOption) (*choreopb.Get_Response, error)
	Apply(ctx context.Context, choreoCtx *choreopb.ChoreoContext, opts ...ApplyOption) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Once(ctx context.Context) (*choreopb.Once_Response, error)
	Load(ctx context.Context) error
	Close() error
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
	client.client = choreopb.NewChoreoClient(conn)
	client.conn = conn
	return client, nil
}

type client struct {
	config *config.Config
	conn   *grpc.ClientConn
	client choreopb.ChoreoClient
}

func (r *client) Close() error {
	return r.conn.Close()
}

func (r *client) Get(ctx context.Context, opts ...GetOption) (*choreopb.Get_Response, error) {
	return r.client.Get(ctx, &choreopb.Get_Request{})
}

func (r *client) Apply(ctx context.Context, choreoCtx *choreopb.ChoreoContext, opts ...ApplyOption) error {
	_, err := r.client.Apply(ctx, &choreopb.Apply_Request{
		ChoreoContext: choreoCtx,
	})
	return err
}

func (r *client) Start(ctx context.Context) error {
	if _, err := r.client.Start(ctx, &choreopb.Start_Request{}); err != nil {
		return err
	}
	return nil
}
func (r *client) Stop(ctx context.Context) error {
	if _, err := r.client.Stop(ctx, &choreopb.Stop_Request{}); err != nil {
		return err
	}
	return nil
}
func (r *client) Once(ctx context.Context) (*choreopb.Once_Response, error) {
	rsp, err := r.client.Once(ctx, &choreopb.Once_Request{})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
func (r *client) Load(ctx context.Context) error {
	if _, err := r.client.Load(ctx, &choreopb.Load_Request{}); err != nil {
		return err
	}
	return nil
}

type GetOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToGet(*GetOptions)
}

var _ GetOption = &GetOptions{}

type GetOptions struct {
	// To be added
}

func (o *GetOptions) ApplyToGet(lo *GetOptions) {
	// To be added
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *GetOptions) ApplyOptions(opts []GetOption) *GetOptions {
	for _, opt := range opts {
		opt.ApplyToGet(o)
	}
	return o
}

type ApplyOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToApply(*ApplyOptions)
}

var _ ApplyOption = &ApplyOptions{}

type ApplyOptions struct {
	// TO be added
}

func (o *ApplyOptions) ApplyToApply(lo *ApplyOptions) {
	// To be added
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *ApplyOptions) ApplyOptions(opts []ApplyOption) *ApplyOptions {
	for _, opt := range opts {
		opt.ApplyToApply(o)
	}
	return o
}

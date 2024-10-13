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
	"k8s.io/apimachinery/pkg/types"
)

type Client interface {
	Get(ctx context.Context, opts ...GetOption) (*choreopb.Get_Response, error)
	Apply(ctx context.Context, choreoCtx *choreopb.ChoreoContext, opts ...ApplyOption) error
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
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *client) Get(ctx context.Context, opts ...GetOption) (*choreopb.Get_Response, error) {
	o := GetOptions{}
	o.ApplyOptions(opts)

	return r.client.Get(ctx, &choreopb.Get_Request{
		Options: &choreopb.Get_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
}

func (r *client) Apply(ctx context.Context, choreoCtx *choreopb.ChoreoContext, opts ...ApplyOption) error {
	o := ApplyOptions{}
	o.ApplyOptions(opts)

	_, err := r.client.Apply(ctx, &choreopb.Apply_Request{
		ChoreoContext: choreoCtx,
		Options: &choreopb.Apply_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	return err
}

type GetOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToGet(*GetOptions)
}

var _ GetOption = &GetOptions{}

type GetOptions struct {
	Proxy types.NamespacedName
}

func (o *GetOptions) ApplyToGet(lo *GetOptions) {
	lo.Proxy = o.Proxy
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
	Proxy types.NamespacedName
}

func (o *ApplyOptions) ApplyToApply(lo *ApplyOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *ApplyOptions) ApplyOptions(opts []ApplyOption) *ApplyOptions {
	for _, opt := range opts {
		opt.ApplyToApply(o)
	}
	return o
}

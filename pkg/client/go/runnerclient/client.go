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
	"k8s.io/apimachinery/pkg/types"
)

type Client interface {
	Start(ctx context.Context, opts ...StartOption) error
	Stop(ctx context.Context, opts ...StopOption) error
	Once(ctx context.Context, opts ...OnceOption) (runnerpb.Runner_OnceClient, error)
	Load(ctx context.Context, opts ...LoadOption) error
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
	client.client = runnerpb.NewRunnerClient(conn)
	client.conn = conn
	return client, nil
}

type client struct {
	config *config.Config
	conn   *grpc.ClientConn
	client runnerpb.RunnerClient
}

func (r *client) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *client) Start(ctx context.Context, opts ...StartOption) error {
	o := StartOptions{}
	o.ApplyOptions(opts)

	if _, err := r.client.Start(ctx, &runnerpb.Start_Request{
		Options: &runnerpb.Start_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	}); err != nil {
		return err
	}
	return nil
}
func (r *client) Stop(ctx context.Context, opts ...StopOption) error {
	o := StopOptions{}
	o.ApplyOptions(opts)

	if _, err := r.client.Stop(ctx, &runnerpb.Stop_Request{
		Options: &runnerpb.Stop_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	}); err != nil {
		return err
	}
	return nil
}
func (r *client) Once(ctx context.Context, opts ...OnceOption) (runnerpb.Runner_OnceClient, error) {
	o := OnceOptions{}
	o.ApplyOptions(opts)

	return r.client.Once(ctx, &runnerpb.Once_Request{
		Options: &runnerpb.Once_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
}
func (r *client) Load(ctx context.Context, opts ...LoadOption) error {
	o := LoadOptions{}
	o.ApplyOptions(opts)

	if _, err := r.client.Load(ctx, &runnerpb.Load_Request{
		Options: &runnerpb.Load_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	}); err != nil {
		return err
	}
	return nil
}

type StartOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToStart(*StartOptions)
}

var _ StartOption = &StartOptions{}

type StartOptions struct {
	Proxy types.NamespacedName
}

func (o *StartOptions) ApplyToStart(lo *StartOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *StartOptions) ApplyOptions(opts []StartOption) *StartOptions {
	for _, opt := range opts {
		opt.ApplyToStart(o)
	}
	return o
}

type StopOption interface {
	ApplyToStop(*StopOptions)
}

var _ StopOption = &StopOptions{}

type StopOptions struct {
	Proxy types.NamespacedName
}

func (o *StopOptions) ApplyToStop(lo *StopOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *StopOptions) ApplyOptions(opts []StopOption) *StopOptions {
	for _, opt := range opts {
		opt.ApplyToStop(o)
	}
	return o
}

type OnceOption interface {
	ApplyToOnce(*OnceOptions)
}

var _ OnceOption = &OnceOptions{}

type OnceOptions struct {
	Proxy types.NamespacedName
}

func (o *OnceOptions) ApplyToOnce(lo *OnceOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *OnceOptions) ApplyOptions(opts []OnceOption) *OnceOptions {
	for _, opt := range opts {
		opt.ApplyToOnce(o)
	}
	return o
}

type LoadOption interface {
	ApplyToLoad(*LoadOptions)
}

var _ LoadOption = &LoadOptions{}

type LoadOptions struct {
	Proxy types.NamespacedName
}

func (o *LoadOptions) ApplyToLoad(lo *LoadOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *LoadOptions) ApplyOptions(opts []LoadOption) *LoadOptions {
	for _, opt := range opts {
		opt.ApplyToLoad(o)
	}
	return o
}

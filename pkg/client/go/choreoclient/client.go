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
	"encoding/json"

	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Client interface {
	Get(ctx context.Context, opts ...GetOption) (*choreopb.Get_Response, error)
	Apply(ctx context.Context, choreoCtx *choreopb.ChoreoContext, opts ...ApplyOption) error
	Start(ctx context.Context, opts ...StartOption) error
	Stop(ctx context.Context, opts ...StopOption) error
	Once(ctx context.Context, opts ...OnceOption) (*choreopb.Once_Response, error)
	Load(ctx context.Context, opts ...LoadOption) error
	List(ctx context.Context, u runtime.Unstructured, opts ...ListOption) error
	Diff(ctx context.Context, opts ...DiffOption) ([]byte, error)
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

func (r *client) Start(ctx context.Context, opts ...StartOption) error {
	o := StartOptions{}
	o.ApplyOptions(opts)

	if _, err := r.client.Start(ctx, &choreopb.Start_Request{
		Options: &choreopb.Start_Options{
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

	if _, err := r.client.Stop(ctx, &choreopb.Stop_Request{
		Options: &choreopb.Stop_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	}); err != nil {
		return err
	}
	return nil
}
func (r *client) Once(ctx context.Context, opts ...OnceOption) (*choreopb.Once_Response, error) {
	o := OnceOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.Once(ctx, &choreopb.Once_Request{
		Options: &choreopb.Once_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
func (r *client) Load(ctx context.Context, opts ...LoadOption) error {
	o := LoadOptions{}
	o.ApplyOptions(opts)

	if _, err := r.client.Load(ctx, &choreopb.Load_Request{
		Options: &choreopb.Load_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	}); err != nil {
		return err
	}
	return nil
}
func (r *client) Diff(ctx context.Context, opts ...DiffOption) ([]byte, error) {
	o := DiffOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.Diff(ctx, &choreopb.Diff_Request{
		Options: &choreopb.Diff_Options{
			ProxyName:        o.Proxy.Name,
			ProxyNamespace:   o.Proxy.Namespace,
			ShowManagedField: o.ShowManagedFields,
			ShowChoreoAPIs:   o.ShowChoreoAPIs,
		},
	})
	if err != nil {
		return nil, err
	}
	//b = rsp.Object
	return rsp.Object, nil

}

func (r *client) List(ctx context.Context, u runtime.Unstructured, opts ...ListOption) error {
	o := ListOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.List(ctx, &choreopb.List_Request{
		Options: &choreopb.List_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return err
	}
	data := map[string]any{}
	if err := json.Unmarshal(rsp.Object, &data); err != nil {
		return err
	}
	u.SetUnstructuredContent(data)
	return nil
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

type DiffOption interface {
	ApplyToDiff(*DiffOptions)
}

var _ DiffOption = &DiffOptions{}

type DiffOptions struct {
	Proxy             types.NamespacedName
	ShowManagedFields bool
	ShowChoreoAPIs    bool
}

func (o *DiffOptions) ApplyToDiff(lo *DiffOptions) {
	lo.Proxy = o.Proxy
	lo.ShowManagedFields = o.ShowManagedFields
	lo.ShowChoreoAPIs = o.ShowChoreoAPIs
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *DiffOptions) ApplyOptions(opts []DiffOption) *DiffOptions {
	for _, opt := range opts {
		opt.ApplyToDiff(o)
	}
	return o
}

type ListOption interface {
	ApplyToList(*ListOptions)
}

var _ ListOption = &ListOptions{}

type ListOptions struct {
	Proxy types.NamespacedName
}

func (o *ListOptions) ApplyToList(lo *ListOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *ListOptions) ApplyOptions(opts []ListOption) *ListOptions {
	for _, opt := range opts {
		opt.ApplyToList(o)
	}
	return o
}

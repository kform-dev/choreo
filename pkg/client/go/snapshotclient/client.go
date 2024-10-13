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

package snapshotclient

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/snapshotpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Client interface {
	Get(ctx context.Context, u runtime.Unstructured, opts ...GetOption) error
	Delete(ctx context.Context, u runtime.Unstructured, opts ...DeleteOption) error
	List(ctx context.Context, u runtime.Unstructured, opts ...ListOption) error
	Diff(ctx context.Context, opts ...DiffOption) ([]byte, error)
	Watch(ctx context.Context, u runtime.Unstructured, opts ...ListOption) chan *snapshotpb.Watch_Response
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
	client.client = snapshotpb.NewSnapshotClient(conn)
	client.conn = conn
	return client, nil
}

type client struct {
	config *config.Config
	conn   *grpc.ClientConn
	client snapshotpb.SnapshotClient
}

func (r *client) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *client) Get(ctx context.Context, u runtime.Unstructured, opts ...GetOption) error {
	o := GetOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.Get(ctx, &snapshotpb.Get_Request{
		Options: &snapshotpb.Get_Options{
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

func (r *client) List(ctx context.Context, u runtime.Unstructured, opts ...ListOption) error {
	o := ListOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.List(ctx, &snapshotpb.List_Request{
		Options: &snapshotpb.List_Options{
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

func (r *client) Delete(ctx context.Context, u runtime.Unstructured, opts ...DeleteOption) error {
	o := DeleteOptions{}
	o.ApplyOptions(opts)

	obj := &unstructured.Unstructured{
		Object: u.UnstructuredContent(),
	}

	_, err := r.client.Delete(ctx, &snapshotpb.Delete_Request{
		Id: obj.GetName(),
		Options: &snapshotpb.Delete_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Diff(ctx context.Context, opts ...DiffOption) ([]byte, error) {
	o := DiffOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.Diff(ctx, &snapshotpb.Diff_Request{
		Options: &snapshotpb.Diff_Options{
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

func (r *client) Watch(ctx context.Context, u runtime.Unstructured, opts ...ListOption) chan *snapshotpb.Watch_Response {
	o := ListOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	var stream snapshotpb.Snapshot_WatchClient
	var err error

	rspCh := make(chan *snapshotpb.Watch_Response)
	go func() {
		defer close(rspCh)
		for {
			select {
			case <-ctx.Done():
				// watch stoppped
				return
			default:
				if stream == nil {
					if stream, err = r.client.Watch(ctx, &snapshotpb.Watch_Request{
						Options: &snapshotpb.Watch_Options{
							ProxyName:      o.Proxy.Name,
							ProxyNamespace: o.Proxy.Namespace,
							Watch:          o.Watch,
						},
					}); err != nil && !errors.Is(err, context.Canceled) {
						if statErr, ok := status.FromError(err); ok {
							switch statErr.Code() {
							case codes.Canceled:
								// dont log when context got cancelled
							default:
								log.Error("watch failed", "error", statErr.Err())
							}
						}
						time.Sleep(time.Second * 1) //- resilience for server crash
						// retry on failure
						continue
					}
				}
				rsp, err := stream.Recv()
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						if er, ok := status.FromError(err); ok {
							switch er.Code() {
							case codes.Canceled:
								log.Debug("resource client watch event recv error", "error", err.Error())
								// dont log when context got cancelled
							default:
								log.Error("failed to receive a message from stream", "error", err.Error())
							}
						}
						// clearing the stream will force the client to resubscribe in the next iteration
						stream = nil
						time.Sleep(time.Second * 1) //- resilience for server crash
						// retry on failure
						continue
					}
					if strings.Contains(err.Error(), "EOF") {
						log.Error("fail rcv", "error", err)
						continue
					}
				}
				log.Debug("resource client event received", "eventType", rsp.EventType.String())
				if rsp.EventType == snapshotpb.Watch_ERROR {
					stream = nil
					time.Sleep(time.Second * 1) //- resilience for server error
					// retry on failure
					continue
				}
				rspCh <- rsp
			}
		}
	}()
	return rspCh
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

type DeleteOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToDelete(*DeleteOptions)
}

var _ DeleteOption = &DeleteOptions{}

type DeleteOptions struct {
	Proxy types.NamespacedName
}

func (o *DeleteOptions) ApplyToDelete(lo *DeleteOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *DeleteOptions) ApplyOptions(opts []DeleteOption) *DeleteOptions {
	for _, opt := range opts {
		opt.ApplyToDelete(o)
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
	Watch bool
}

func (o *ListOptions) ApplyToList(lo *ListOptions) {
	lo.Proxy = o.Proxy
	lo.Watch = o.Watch
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *ListOptions) ApplyOptions(opts []ListOption) *ListOptions {
	for _, opt := range opts {
		opt.ApplyToList(o)
	}
	return o
}

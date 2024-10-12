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

package branchclient

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
)

type Client interface {
	Get(ctx context.Context, branch string, opt ...GetOption) ([]*branchpb.Get_Log, error)
	List(ctx context.Context, opt ...ListOption) ([]*branchpb.BranchObject, error)
	Create(ctx context.Context, branch string, opt ...CreateOption) error
	Delete(ctx context.Context, branch string, opt ...DeleteOption) error
	Diff(ctx context.Context, srcbranch, dstbranch string, opt ...DiffOption) ([]*branchpb.Diff_Diff, error)
	Merge(ctx context.Context, srcbranch, dstbranch string, opt ...MergeOption) error
	Stash(ctx context.Context, branch string, opt ...StashOption) error
	Checkout(ctx context.Context, branch string, opt ...CheckoutOption) error
	StreamFiles(ctx context.Context, branch string, opts ...ListOption) chan *branchpb.Get_File
	Watch(ctx context.Context, in *branchpb.Watch_Request, opts ...ListOption) chan *branchpb.Watch_Response
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
	client.client = branchpb.NewBranchClient(conn)
	client.conn = conn
	return client, nil
}

type client struct {
	config *config.Config
	conn   *grpc.ClientConn
	client branchpb.BranchClient
}

func (r *client) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *client) Get(ctx context.Context, branch string, opts ...GetOption) ([]*branchpb.Get_Log, error) {
	o := GetOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.Get(ctx, &branchpb.Get_Request{
		Branch: branch,
		Options: &branchpb.Get_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return nil, err
	}
	return rsp.GetLogs(), nil
}

func (r *client) List(ctx context.Context, opts ...ListOption) ([]*branchpb.BranchObject, error) {
	o := ListOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.List(ctx, &branchpb.List_Request{
		Options: &branchpb.List_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return nil, err
	}
	return rsp.BranchObjects, nil
}

func (r *client) Create(ctx context.Context, branch string, opts ...CreateOption) error {
	o := CreateOptions{}
	o.ApplyOptions(opts)

	_, err := r.client.Create(ctx, &branchpb.Create_Request{
		Branch: branch,
		Options: &branchpb.Create_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Delete(ctx context.Context, branch string, opts ...DeleteOption) error {
	o := DeleteOptions{}
	o.ApplyOptions(opts)

	_, err := r.client.Delete(ctx, &branchpb.Delete_Request{
		Branch: branch,
		Options: &branchpb.Delete_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Diff(ctx context.Context, srcbranch, dstbranch string, opts ...DiffOption) ([]*branchpb.Diff_Diff, error) {
	o := DiffOptions{}
	o.ApplyOptions(opts)

	rsp, err := r.client.Diff(ctx, &branchpb.Diff_Request{
		SrcBranch: srcbranch,
		DstBranch: dstbranch,
		Options: &branchpb.Diff_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return nil, err
	}
	return rsp.Diffs, nil
}

func (r *client) Merge(ctx context.Context, srcbranch, dstbranch string, opts ...MergeOption) error {
	o := MergeOptions{}
	o.ApplyOptions(opts)

	_, err := r.client.Merge(ctx, &branchpb.Merge_Request{
		SrcBranch: srcbranch,
		DstBranch: dstbranch,
		Options: &branchpb.Merge_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Stash(ctx context.Context, branch string, opts ...StashOption) error {
	o := StashOptions{}
	o.ApplyOptions(opts)

	_, err := r.client.Stash(ctx, &branchpb.Stash_Request{
		Branch: branch,
		Options: &branchpb.Stash_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Checkout(ctx context.Context, branch string, opts ...CheckoutOption) error {
	o := CheckoutOptions{}
	o.ApplyOptions(opts)

	_, err := r.client.Checkout(ctx, &branchpb.Checkout_Request{
		Branch: branch,
		Options: &branchpb.Checkout_Options{
			ProxyName:      o.Proxy.Name,
			ProxyNamespace: o.Proxy.Namespace,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) StreamFiles(ctx context.Context, branch string, opts ...ListOption) chan *branchpb.Get_File {
	o := ListOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	var stream branchpb.Branch_StreamFilesClient
	var err error

	rspCh := make(chan *branchpb.Get_File)
	go func() {
		defer close(rspCh)
		ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
		defer cancel()
		stream, err = r.client.StreamFiles(ctx, &branchpb.Get_Request{
			Branch: branch,
			Options: &branchpb.Get_Options{
				ProxyName:      o.Proxy.Name,
				ProxyNamespace: o.Proxy.Namespace,
			},
		})
		if err != nil {
			log.Error("failed to get stream", "error", err)
			return
		}
		for {
			rsp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return // End of stream
				}
				log.Error("stream rsp error", "error", err)
				return
			}
			select {
			case <-ctx.Done():
				// Context cancellation or deadline exceeded
				log.Debug("stream context cancelled", "reason", ctx.Err())
				return
			default:
				rspCh <- rsp
			}
		}
	}()
	return rspCh
}

func (r *client) Watch(ctx context.Context, in *branchpb.Watch_Request, opts ...ListOption) chan *branchpb.Watch_Response {
	o := ListOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	var stream branchpb.Branch_WatchClient
	var err error

	rspCh := make(chan *branchpb.Watch_Response)
	go func() {
		defer close(rspCh)
		for {
			select {
			case <-ctx.Done():
				// watch stoppped
				return
			default:
				if stream == nil {
					if stream, err = r.client.Watch(ctx, in); err != nil && !errors.Is(err, context.Canceled) {
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
				if rsp.EventType == branchpb.Watch_ERROR {
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

type ListOption interface {
	// ApplyToGet applies this configuration to the given get options.
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

type CreateOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToCreate(*CreateOptions)
}

var _ CreateOption = &CreateOptions{}

type CreateOptions struct {
	Proxy types.NamespacedName
}

func (o *CreateOptions) ApplyToCreate(lo *CreateOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *CreateOptions) ApplyOptions(opts []CreateOption) *CreateOptions {
	for _, opt := range opts {
		opt.ApplyToCreate(o)
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

type DiffOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToDiff(*DiffOptions)
}

var _ DiffOption = &DiffOptions{}

type DiffOptions struct {
	Proxy types.NamespacedName
}

func (o *DiffOptions) ApplyToDiff(lo *DiffOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *DiffOptions) ApplyOptions(opts []DiffOption) *DiffOptions {
	for _, opt := range opts {
		opt.ApplyToDiff(o)
	}
	return o
}

type MergeOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToMerge(*MergeOptions)
}

var _ MergeOption = &MergeOptions{}

type MergeOptions struct {
	Proxy types.NamespacedName
}

func (o *MergeOptions) ApplyToMerge(lo *MergeOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *MergeOptions) ApplyOptions(opts []MergeOption) *MergeOptions {
	for _, opt := range opts {
		opt.ApplyToMerge(o)
	}
	return o
}

type StashOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToStash(*StashOptions)
}

var _ StashOption = &StashOptions{}

type StashOptions struct {
	Proxy types.NamespacedName
}

func (o *StashOptions) ApplyToStash(lo *StashOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *StashOptions) ApplyOptions(opts []StashOption) *StashOptions {
	for _, opt := range opts {
		opt.ApplyToStash(o)
	}
	return o
}

type CheckoutOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToCheckout(*CheckoutOptions)
}

var _ CheckoutOption = &CheckoutOptions{}

type CheckoutOptions struct {
	Proxy types.NamespacedName
}

func (o *CheckoutOptions) ApplyToCheckout(lo *CheckoutOptions) {
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *CheckoutOptions) ApplyOptions(opts []CheckoutOption) *CheckoutOptions {
	for _, opt := range opts {
		opt.ApplyToCheckout(o)
	}
	return o
}

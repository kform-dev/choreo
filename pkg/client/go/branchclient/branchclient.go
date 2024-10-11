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
)

type BranchClient interface {
	Get(ctx context.Context, in *branchpb.Get_Request, opts ...grpc.CallOption) (*branchpb.Get_Response, error)
	List(ctx context.Context, in *branchpb.List_Request, opts ...grpc.CallOption) (*branchpb.List_Response, error)
	Create(ctx context.Context, in *branchpb.Create_Request, opts ...grpc.CallOption) (*branchpb.Create_Response, error)
	Delete(ctx context.Context, in *branchpb.Delete_Request, opts ...grpc.CallOption) (*branchpb.Delete_Response, error)
	Merge(ctx context.Context, in *branchpb.Merge_Request, opts ...grpc.CallOption) (*branchpb.Merge_Response, error)
	Diff(ctx context.Context, in *branchpb.Diff_Request, opts ...grpc.CallOption) (*branchpb.Diff_Response, error)
	Stash(ctx context.Context, in *branchpb.Stash_Request, opts ...grpc.CallOption) (*branchpb.Stash_Response, error)
	Checkout(ctx context.Context, in *branchpb.Checkout_Request, opts ...grpc.CallOption) (*branchpb.Checkout_Response, error)
	StreamFiles(ctx context.Context, in *branchpb.Get_Request, opts ...grpc.CallOption) chan *branchpb.Get_File
	Watch(ctx context.Context, in *branchpb.Watch_Request, opts ...grpc.CallOption) chan *branchpb.Watch_Response
	Close() error
}

func NewBranchClient(config *config.Config) (BranchClient, error) {
	client := &branchclient{
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

type branchclient struct {
	config *config.Config
	conn   *grpc.ClientConn
	client branchpb.BranchClient
}

func (r *branchclient) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *branchclient) Get(ctx context.Context, in *branchpb.Get_Request, opts ...grpc.CallOption) (*branchpb.Get_Response, error) {
	return r.client.Get(ctx, in, opts...)
}
func (r *branchclient) List(ctx context.Context, in *branchpb.List_Request, opts ...grpc.CallOption) (*branchpb.List_Response, error) {
	return r.client.List(ctx, in, opts...)
}
func (r *branchclient) Create(ctx context.Context, in *branchpb.Create_Request, opts ...grpc.CallOption) (*branchpb.Create_Response, error) {
	return r.client.Create(ctx, in, opts...)
}
func (r *branchclient) Delete(ctx context.Context, in *branchpb.Delete_Request, opts ...grpc.CallOption) (*branchpb.Delete_Response, error) {
	return r.client.Delete(ctx, in, opts...)
}
func (r *branchclient) Merge(ctx context.Context, in *branchpb.Merge_Request, opts ...grpc.CallOption) (*branchpb.Merge_Response, error) {
	return r.client.Merge(ctx, in, opts...)
}
func (r *branchclient) Diff(ctx context.Context, in *branchpb.Diff_Request, opts ...grpc.CallOption) (*branchpb.Diff_Response, error) {
	return r.client.Diff(ctx, in, opts...)
}
func (r *branchclient) Stash(ctx context.Context, in *branchpb.Stash_Request, opts ...grpc.CallOption) (*branchpb.Stash_Response, error) {
	return r.client.Stash(ctx, in, opts...)
}
func (r *branchclient) Checkout(ctx context.Context, in *branchpb.Checkout_Request, opts ...grpc.CallOption) (*branchpb.Checkout_Response, error) {
	return r.client.Checkout(ctx, in, opts...)
}
func (r *branchclient) StreamFiles(ctx context.Context, in *branchpb.Get_Request, opts ...grpc.CallOption) chan *branchpb.Get_File {
	log := log.FromContext(ctx)
	var stream branchpb.Branch_StreamFilesClient
	var err error

	rspCh := make(chan *branchpb.Get_File)
	go func() {
		defer close(rspCh)
		ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
		defer cancel()
		stream, err = r.client.StreamFiles(ctx, in)
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

func (r *branchclient) Watch(ctx context.Context, in *branchpb.Watch_Request, opts ...grpc.CallOption) chan *branchpb.Watch_Response {
	log := log.FromContext(ctx)
	var stream branchpb.Branch_WatchClient
	var err error

	rspCh := make(chan *branchpb.Watch_Response)
	go func() {
		defer close(rspCh)
		for {
			select {
			case <-ctx.Done():
				log.Info("watch stopped", "client", "proxy")
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
								log.Info("resource client watch event recv error", "error", err.Error())
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
						log.Error("fail rcv", "err", err)
						continue
					}
				}
				log.Info("resource client event received", "eventType", rsp.EventType)
				rspCh <- rsp
			}
		}
	}()
	return rspCh
}

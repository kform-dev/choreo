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

package resourceclient

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type ResourceClient interface {
	Get(ctx context.Context, in *resourcepb.Get_Request, opts ...grpc.CallOption) (*resourcepb.Get_Response, error)
    List(ctx context.Context, in *resourcepb.List_Request, opts ...grpc.CallOption) (*resourcepb.List_Response, error)
    Create(ctx context.Context, in *resourcepb.Create_Request, opts ...grpc.CallOption) (*resourcepb.Create_Response, error)
    Update(ctx context.Context, in *resourcepb.Update_Request, opts ...grpc.CallOption) (*resourcepb.Update_Response, error)
    Apply(ctx context.Context, in *resourcepb.Apply_Request, opts ...grpc.CallOption) (*resourcepb.Apply_Response, error)
    Delete(ctx context.Context, in *resourcepb.Delete_Request, opts ...grpc.CallOption) (*resourcepb.Delete_Response, error)
    Watch(ctx context.Context, in *resourcepb.Watch_Request, opts ...grpc.CallOption) chan *resourcepb.Watch_Response
	Close() error
}

func NewResourceClient(config *config.Config) (ResourceClient, error) {
	client := &resourceclient{
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
	client.client = resourcepb.NewResourceClient(conn)
	client.conn = conn
	return client, nil
}

type resourceclient struct {
	config *config.Config
	conn   *grpc.ClientConn
	client resourcepb.ResourceClient
}

func (r *resourceclient) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *resourceclient) Get(ctx context.Context, in *resourcepb.Get_Request, opts ...grpc.CallOption) (*resourcepb.Get_Response, error) {
	return r.client.Get(ctx, in, opts...)
}
func (r *resourceclient) List(ctx context.Context, in *resourcepb.List_Request, opts ...grpc.CallOption) (*resourcepb.List_Response, error) {
	return r.client.List(ctx, in, opts...)
}
func (r *resourceclient) Create(ctx context.Context, in *resourcepb.Create_Request, opts ...grpc.CallOption) (*resourcepb.Create_Response, error) {
	return r.client.Create(ctx, in, opts...)
}
func (r *resourceclient) Update(ctx context.Context, in *resourcepb.Update_Request, opts ...grpc.CallOption) (*resourcepb.Update_Response, error) {
	return r.client.Update(ctx, in, opts...)
}
func (r *resourceclient) Apply(ctx context.Context, in *resourcepb.Apply_Request, opts ...grpc.CallOption) (*resourcepb.Apply_Response, error) {
	return r.client.Apply(ctx, in, opts...)
}
func (r *resourceclient) Delete(ctx context.Context, in *resourcepb.Delete_Request, opts ...grpc.CallOption) (*resourcepb.Delete_Response, error) {
	return r.client.Delete(ctx, in, opts...)
}
func (r *resourceclient) Watch(ctx context.Context, in *resourcepb.Watch_Request, opts ...grpc.CallOption) chan *resourcepb.Watch_Response {
	log := log.FromContext(ctx)
	var stream resourcepb.Resource_WatchClient
	var err error

	rspCh := make(chan *resourcepb.Watch_Response)
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

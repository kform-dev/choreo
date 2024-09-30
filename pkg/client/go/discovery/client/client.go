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

package client

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type DiscoveryClientInterface interface {
	Get(context.Context, *discoverypb.Get_Request) (*discoverypb.Get_Response, error)
	Watch(context.Context, *discoverypb.Watch_Request) chan *discoverypb.Watch_Response
	Close() error
}

func New(config *config.Config) (DiscoveryClientInterface, error) {
	conn, err := grpc.NewClient(config.Address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(config.MaxMsgSize)),
	)
	if err != nil {
		return nil, err
	}
	client := discoverypb.NewDiscoveryClient(conn)
	return &discoveryclient{
		conn:   conn,
		client: client,
		config: config,
	}, nil
}

type discoveryclient struct {
	config *config.Config
	conn   *grpc.ClientConn
	client discoverypb.DiscoveryClient
}

// Close cleanly closes the gRPC connection when the client is no longer needed.
func (r *discoveryclient) Close() error {
	return r.conn.Close()
}

func (r *discoveryclient) Get(ctx context.Context, in *discoverypb.Get_Request) (*discoverypb.Get_Response, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	return r.client.Get(ctx, in)
}

func (r *discoveryclient) Watch(ctx context.Context, in *discoverypb.Watch_Request) chan *discoverypb.Watch_Response {
	log := log.FromContext(ctx)
	var stream discoverypb.Discovery_WatchClient
	var err error

	rspCh := make(chan *discoverypb.Watch_Response)
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
				if rsp.EventType == discoverypb.Watch_ERROR {
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

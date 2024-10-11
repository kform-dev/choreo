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

type GetOptions struct{}

type ListOptions struct{
	Choreo string
}

type CreateOptions struct{}

type DeleteOptions struct{}

type DiffOptions struct{}

type MergeOptions struct{}

type StashOptions struct{}

type CheckoutOptions struct{}

type Client interface {
	Get(ctx context.Context, branch string, opt GetOptions) ([]*branchpb.Get_Log, error)
	List(ctx context.Context, opt ListOptions) ([]*branchpb.BranchObject, error)
	Create(ctx context.Context, branch string, opt CreateOptions) error
	Delete(ctx context.Context, branch string, opt DeleteOptions) error
	Diff(ctx context.Context, srcbranch, dstbranch string, opt DiffOptions) ([]*branchpb.Diff_Diff, error)
	Merge(ctx context.Context, srcbranch, dstbranch string, opt MergeOptions) error
	Stash(ctx context.Context, branch string, opt StashOptions) error
	Checkout(ctx context.Context, branch string, opt CheckoutOptions) error
	StreamFiles(ctx context.Context, branch string) chan *branchpb.Get_File
	Watch(ctx context.Context, in *branchpb.Watch_Request, opts ...grpc.CallOption) chan *branchpb.Watch_Response
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

func (r *client) Get(ctx context.Context, branch string, opt GetOptions) ([]*branchpb.Get_Log, error) {
	rsp, err := r.client.Get(ctx, &branchpb.Get_Request{
		Branch:  branch,
		Options: &branchpb.Get_Options{},
	})
	if err != nil {
		return nil, err
	}
	return rsp.GetLogs(), nil
}

func (r *client) List(ctx context.Context, opt ListOptions) ([]*branchpb.BranchObject, error) {
	rsp, err := r.client.List(ctx, &branchpb.List_Request{
		Options: &branchpb.List_Options{},
	})
	if err != nil {
		return nil, err
	}
	return rsp.BranchObjects, nil
}

func (r *client) Create(ctx context.Context, branch string, opt CreateOptions) error {
	_, err := r.client.Create(ctx, &branchpb.Create_Request{
		Branch:  branch,
		Options: &branchpb.Create_Options{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Delete(ctx context.Context, branch string, opt DeleteOptions) error {
	_, err := r.client.Delete(ctx, &branchpb.Delete_Request{
		Branch:  branch,
		Options: &branchpb.Delete_Options{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Diff(ctx context.Context, srcbranch, dstbranch string, opt DiffOptions) ([]*branchpb.Diff_Diff, error) {
	rsp, err := r.client.Diff(ctx, &branchpb.Diff_Request{
		SrcBranch: srcbranch,
		DstBranch: dstbranch,
		Options:   &branchpb.Diff_Options{},
	})
	if err != nil {
		return nil, err
	}
	return rsp.Diffs, nil
}

func (r *client) Merge(ctx context.Context, srcbranch, dstbranch string, opt MergeOptions) error {
	_, err := r.client.Merge(ctx, &branchpb.Merge_Request{
		SrcBranch: srcbranch,
		DstBranch: dstbranch,
		Options:   &branchpb.Merge_Options{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Stash(ctx context.Context, branch string, opt StashOptions) error {
	_, err := r.client.Stash(ctx, &branchpb.Stash_Request{
		Branch:  branch,
		Options: &branchpb.Stash_Options{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Checkout(ctx context.Context, branch string, opt CheckoutOptions) error {
	_, err := r.client.Checkout(ctx, &branchpb.Checkout_Request{
		Branch:  branch,
		Options: &branchpb.Checkout_Options{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) StreamFiles(ctx context.Context, branch string) chan *branchpb.Get_File {
	log := log.FromContext(ctx)
	var stream branchpb.Branch_StreamFilesClient
	var err error

	rspCh := make(chan *branchpb.Get_File)
	go func() {
		defer close(rspCh)
		ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
		defer cancel()
		stream, err = r.client.StreamFiles(ctx, &branchpb.Get_Request{Branch: branch})
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

func (r *client) Watch(ctx context.Context, in *branchpb.Watch_Request, opts ...grpc.CallOption) chan *branchpb.Watch_Response {
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

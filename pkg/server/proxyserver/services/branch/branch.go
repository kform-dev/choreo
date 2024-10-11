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

package branch

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/kform-dev/choreo/pkg/server/proxyserver/choreoctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
)

func New(store store.Storer[*choreoctx.ChoreoCtx]) branchpb.BranchServer {
	return &proxy{
		store: store,
	}
}

type proxy struct {
	branchpb.UnimplementedBranchServer
	store store.Storer[*choreoctx.ChoreoCtx]
}

func (r *proxy) getChoreoCtx(nsn types.NamespacedName) (*choreoctx.ChoreoCtx, error) {
	choreoCtx, err := r.store.Get(store.KeyFromNSN(nsn))
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("choreo %s not found, err: %v", nsn.String(), err))
	}
	if !choreoCtx.Ready {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("choreo %s not ready, err: %v", nsn.String(), err))
	}
	return choreoCtx, nil
}

func (r *proxy) Get(ctx context.Context, req *branchpb.Get_Request) (*branchpb.Get_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.Get_Response{}, err
	}
	return choreoCtx.BranchClient.Get(ctx, req)
}

func (r *proxy) List(ctx context.Context, req *branchpb.List_Request) (*branchpb.List_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.List_Response{}, err
	}
	return choreoCtx.BranchClient.List(ctx, req)
}

func (r *proxy) Create(ctx context.Context, req *branchpb.Create_Request) (*branchpb.Create_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.Create_Response{}, err
	}
	return choreoCtx.BranchClient.Create(ctx, req)
}

func (r *proxy) Delete(ctx context.Context, req *branchpb.Delete_Request) (*branchpb.Delete_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.Delete_Response{}, err
	}
	return choreoCtx.BranchClient.Delete(ctx, req)
}

func (r *proxy) Diff(ctx context.Context, req *branchpb.Diff_Request) (*branchpb.Diff_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.Diff_Response{}, err
	}
	return choreoCtx.BranchClient.Diff(ctx, req)
}

func (r *proxy) Merge(ctx context.Context, req *branchpb.Merge_Request) (*branchpb.Merge_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.Merge_Response{}, err
	}
	return choreoCtx.BranchClient.Merge(ctx, req)
}

func (r *proxy) Stash(ctx context.Context, req *branchpb.Stash_Request) (*branchpb.Stash_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.Stash_Response{}, err
	}
	return choreoCtx.BranchClient.Stash(ctx, req)
}

func (r *proxy) Checkout(ctx context.Context, req *branchpb.Checkout_Request) (*branchpb.Checkout_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return &branchpb.Checkout_Response{}, err
	}
	return choreoCtx.BranchClient.Checkout(ctx, req)
}

func (r *proxy) StreamFiles(req *branchpb.Get_Request, stream branchpb.Branch_StreamFilesServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)
	log.Info("watch")
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go r.streamFiles(ctx, choreoCtx.BranchClient, req, stream)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Info("watch stopped, program cancelled")
	return nil
}

func (r *proxy) streamFiles(ctx context.Context, client branchclient.BranchClient, req *branchpb.Get_Request, clientStream branchpb.Branch_StreamFilesServer) {
	log := log.FromContext(ctx)
	// start watching
	rspCh := client.StreamFiles(ctx, req)

	for {
		select {
		case <-ctx.Done():
			log.Info("stream files stopped", "client", "proxy")
			return
		case rsp, ok := <-rspCh:
			if !ok {
				log.Info("watch done")
				return
			}
			if err := clientStream.Send(rsp); err != nil {
				p, _ := peer.FromContext(clientStream.Context())
				addr := "unknown"
				if p != nil {
					addr = p.Addr.String()
				}
				log.Error("proxy send stream failed", "client", addr)
			}
		}
	}
}

func (r *proxy) Watch(req *branchpb.Watch_Request, stream branchpb.Branch_WatchServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)
	log.Info("watch")
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Name: req.Choreo, Namespace: "default"})
	if err != nil {
		return err
	}

	req.Id = uuid.New().String()

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go r.watch(ctx, choreoCtx.BranchClient, req, stream)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Info("watch stopped, program cancelled")
	return nil
}

func (r *proxy) watch(ctx context.Context, client branchclient.BranchClient, req *branchpb.Watch_Request, clientStream branchpb.Branch_WatchServer) {
	log := log.FromContext(ctx)
	// start watching
	rspCh := client.Watch(ctx, req)

	for {
		select {
		case <-ctx.Done():
			log.Info("watch stopped", "client", "proxy")
			return
		case rsp, ok := <-rspCh:
			if !ok {
				log.Info("watch done")
				return
			}
			if err := clientStream.Send(rsp); err != nil {
				p, _ := peer.FromContext(clientStream.Context())
				addr := "unknown"
				if p != nil {
					addr = p.Addr.String()
				}
				log.Error("proxy send stream failed", "client", addr)
			}
		}
	}
}

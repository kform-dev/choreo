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

package resource

import (
	"context"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/proxyserver/choreoctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
)

func New(store store.Storer[*choreoctx.ChoreoCtx]) resourcepb.ResourceServer {
	return &proxy{
		store: store,
	}
}

type proxy struct {
	resourcepb.UnimplementedResourceServer
	store store.Storer[*choreoctx.ChoreoCtx]
}

func (r *proxy) getChoreoCtx(proxy types.NamespacedName) (*choreoctx.ChoreoCtx, error) {
	choreoCtx, err := r.store.Get(store.KeyFromNSN(proxy))
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("choreo %s not found, err: %v", proxy.String(), err))
	}
	if !choreoCtx.Ready {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("choreo %s not ready, err: %v", proxy.String(), err))
	}
	return choreoCtx, nil
}

func (r *proxy) Get(ctx context.Context, req *resourcepb.Get_Request) (*resourcepb.Get_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &resourcepb.Get_Response{}, err
	}
	rsp, err := choreoCtx.ResourceClient.Get(ctx, req)
	return rsp, err
}

func (r *proxy) List(ctx context.Context, req *resourcepb.List_Request) (*resourcepb.List_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &resourcepb.List_Response{}, err
	}
	rsp, err := choreoCtx.ResourceClient.List(ctx, req)
	return rsp, err
}

func (r *proxy) Apply(ctx context.Context, req *resourcepb.Apply_Request) (*resourcepb.Apply_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &resourcepb.Apply_Response{}, err
	}
	return choreoCtx.ResourceClient.Apply(ctx, req)
}

func (r *proxy) Create(ctx context.Context, req *resourcepb.Create_Request) (*resourcepb.Create_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &resourcepb.Create_Response{}, err
	}
	return choreoCtx.ResourceClient.Create(ctx, req)
}

func (r *proxy) Update(ctx context.Context, req *resourcepb.Update_Request) (*resourcepb.Update_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &resourcepb.Update_Response{}, err
	}
	return choreoCtx.ResourceClient.Update(ctx, req)
}

func (r *proxy) Delete(ctx context.Context, req *resourcepb.Delete_Request) (*resourcepb.Delete_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &resourcepb.Delete_Response{}, err
	}
	return choreoCtx.ResourceClient.Delete(ctx, req)
}

func (r *proxy) Watch(req *resourcepb.Watch_Request, stream resourcepb.Resource_WatchServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)
	log.Info("watch")
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go r.watch(ctx, choreoCtx.ResourceClient, req, stream)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Info("watch stopped, program cancelled")
	return nil
}

func (r *proxy) watch(ctx context.Context, client resourceclient.ResourceClient, req *resourcepb.Watch_Request, clientStream resourcepb.Resource_WatchServer) {
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

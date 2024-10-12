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

package discovery

import (
	"context"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/client/go/discoveryclient"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/server/proxyserver/choreoctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
)

func New(store store.Storer[*choreoctx.ChoreoCtx]) discoverypb.DiscoveryServer {
	return &proxy{
		store: store,
	}
}

type proxy struct {
	discoverypb.UnimplementedDiscoveryServer
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

func (r *proxy) Get(ctx context.Context, req *discoverypb.Get_Request) (*discoverypb.Get_Response, error) {
	fmt.Println("proxy get", req.ProxyName, req.ProxyNamespace)
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.ProxyNamespace, Name: req.ProxyName})
	if err != nil {
		return &discoverypb.Get_Response{}, err
	}

	return choreoCtx.DiscoveryClient.Get(ctx, req)
}

func (r *proxy) Watch(req *discoverypb.Watch_Request, stream discoverypb.Discovery_WatchServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)
	log.Info("watch")
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.ProxyNamespace, Name: req.ProxyName})
	if err != nil {
		return err
	}

	//req.Id = uuid.New().String()

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go r.watch(ctx, choreoCtx.DiscoveryClient, req, stream)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Info("watch stopped, program cancelled")
	return nil
}

func (r *proxy) watch(ctx context.Context, client discoveryclient.DiscoveryClient, req *discoverypb.Watch_Request, clientStream discoverypb.Discovery_WatchServer) {
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

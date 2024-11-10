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

package runner

import (
	"context"
	"fmt"
	"io"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/server/proxyserver/choreoctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
)

func New(store store.Storer[*choreoctx.ChoreoCtx]) runnerpb.RunnerServer {
	return &proxy{
		store: store,
	}
}

type proxy struct {
	runnerpb.UnimplementedRunnerServer
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

func (r *proxy) Start(ctx context.Context, req *runnerpb.Start_Request) (*runnerpb.Start_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &runnerpb.Start_Response{}, err
	}
	return choreoCtx.RunnerClient.Start(ctx, req)
}

func (r *proxy) Stop(ctx context.Context, req *runnerpb.Stop_Request) (*runnerpb.Stop_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &runnerpb.Stop_Response{}, err
	}
	return choreoCtx.RunnerClient.Stop(ctx, req)
}

func (r *proxy) Load(ctx context.Context, req *runnerpb.Load_Request) (*runnerpb.Load_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &runnerpb.Load_Response{}, err
	}
	return choreoCtx.RunnerClient.Load(ctx, req)
}

func (r *proxy) Once(req *runnerpb.Once_Request, stream runnerpb.Runner_OnceServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)

	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go r.once(ctx, choreoCtx.RunnerClient, req, stream)
	<-ctx.Done()
	log.Debug("watch stopped, program cancelled")
	return nil

}

func (r *proxy) once(ctx context.Context, client runnerclient.RunnerClient, req *runnerpb.Once_Request, clientStream runnerpb.Runner_OnceServer) {
	log := log.FromContext(ctx)
	// start watching
	stream, err := client.Once(ctx, req)
	if err != nil {
		log.Error(err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("stream files stopped", "client", "proxy")
			return
		default:
			rsp, err := stream.Recv()
			if err == io.EOF {
				break // Stream is closed by the server
			}
			if err != nil {
				log.Error(err.Error())
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

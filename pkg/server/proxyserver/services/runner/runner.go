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

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/server/proxyserver/choreoctx"
	"google.golang.org/grpc/codes"
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

func (r *proxy) Once(ctx context.Context, req *runnerpb.Once_Request) (*runnerpb.Once_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &runnerpb.Once_Response{}, err
	}
	return choreoCtx.RunnerClient.Once(ctx, req)
}
func (r *proxy) Load(ctx context.Context, req *runnerpb.Load_Request) (*runnerpb.Load_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &runnerpb.Load_Response{}, err
	}
	return choreoCtx.RunnerClient.Load(ctx, req)
}

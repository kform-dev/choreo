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

package choreo

import (
	"context"
	"fmt"

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/server/proxyserver/choreoctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
)

func New(store store.Storer[*choreoctx.ChoreoCtx]) choreopb.ChoreoServer {
	return &proxy{
		store: store,
	}
}

type proxy struct {
	choreopb.UnimplementedChoreoServer
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

func (r *proxy) Get(ctx context.Context, req *choreopb.Get_Request) (*choreopb.Get_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &choreopb.Get_Response{}, err
	}

	return choreoCtx.ChoreoClient.Get(ctx, req)
}

func (r *proxy) Apply(ctx context.Context, req *choreopb.Apply_Request) (*choreopb.Apply_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &choreopb.Apply_Response{}, err
	}

	return choreoCtx.ChoreoClient.Apply(ctx, req)
}

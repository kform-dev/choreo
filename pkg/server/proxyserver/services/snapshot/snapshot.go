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

package snapshot

import (
	"context"
	"fmt"

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/proto/snapshotpb"
	"github.com/kform-dev/choreo/pkg/server/proxyserver/choreoctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
)

func New(store store.Storer[*choreoctx.ChoreoCtx]) snapshotpb.SnapshotServer {
	return &proxy{
		store: store,
	}
}

type proxy struct {
	snapshotpb.UnimplementedSnapshotServer
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

func (r *proxy) Get(ctx context.Context, req *snapshotpb.Get_Request) (*snapshotpb.Get_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &snapshotpb.Get_Response{}, err
	}
	return choreoCtx.SnapshotClient.Get(ctx, req)
}

func (r *proxy) List(ctx context.Context, req *snapshotpb.List_Request) (*snapshotpb.List_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &snapshotpb.List_Response{}, err
	}
	return choreoCtx.SnapshotClient.List(ctx, req)
}

func (r *proxy) Delete(ctx context.Context, req *snapshotpb.Delete_Request) (*snapshotpb.Delete_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &snapshotpb.Delete_Response{}, err
	}
	return choreoCtx.SnapshotClient.Delete(ctx, req)
}

func (r *proxy) Diff(ctx context.Context, req *snapshotpb.Diff_Request) (*snapshotpb.Diff_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &snapshotpb.Diff_Response{}, err
	}
	return choreoCtx.SnapshotClient.Diff(ctx, req)
}

func (r *proxy) Result(ctx context.Context, req *snapshotpb.Result_Request) (*snapshotpb.Result_Response, error) {
	choreoCtx, err := r.getChoreoCtx(types.NamespacedName{Namespace: req.Options.ProxyNamespace, Name: req.Options.ProxyName})
	if err != nil {
		return &snapshotpb.Result_Response{}, err
	}
	return choreoCtx.SnapshotClient.Result(ctx, req)
}

// TODO implement watch

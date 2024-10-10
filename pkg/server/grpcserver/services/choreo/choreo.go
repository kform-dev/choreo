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

	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func New(choreo choreo.Choreo) choreopb.ChoreoServer {
	return &srv{
		choreo: choreo,
	}
}

type srv struct {
	choreopb.UnimplementedChoreoServer
	choreo choreo.Choreo
}

func (r *srv) Get(ctx context.Context, req *choreopb.Get_Request) (*choreopb.Get_Response, error) {
	return r.choreo.Get(ctx, req)
}

func (r *srv) Apply(ctx context.Context, req *choreopb.Apply_Request) (*choreopb.Apply_Response, error) {
	return r.choreo.Apply(ctx, req)
}

func (r *srv) Start(ctx context.Context, req *choreopb.Start_Request) (*choreopb.Start_Response, error) {
	bctx, err := r.choreo.GetBranchStore().GetCheckedOut()
	if bctx == nil {
		return nil, status.Errorf(codes.NotFound, "no checkedout branch found %v", err)
	}
	return r.choreo.Runner().Start(ctx, bctx)
}

func (r *srv) Stop(ctx context.Context, req *choreopb.Stop_Request) (*choreopb.Stop_Response, error) {
	r.choreo.Runner().Stop()
	return &choreopb.Stop_Response{}, nil
}

func (r *srv) Once(ctx context.Context, req *choreopb.Once_Request) (*choreopb.Once_Response, error) {
	bctx, err := r.choreo.GetBranchStore().GetCheckedOut()
	if bctx == nil {
		return nil, status.Errorf(codes.NotFound, "no checkedout branch found %v", err)
	}

	return r.choreo.Runner().RunOnce(ctx, bctx)
}
func (r *srv) Load(ctx context.Context, req *choreopb.Load_Request) (*choreopb.Load_Response, error) {
	bctx, err := r.choreo.GetBranchStore().GetCheckedOut()
	if bctx == nil {
		return nil, status.Errorf(codes.NotFound, "no checkedout branch found %v", err)
	}
	if err := bctx.State.LoadData(ctx, bctx); err != nil {
		return nil, status.Errorf(codes.Internal, "load data failed %v", err)
	}

	return &choreopb.Load_Response{}, nil
}

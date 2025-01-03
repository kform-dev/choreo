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

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func New(choreo choreo.Choreo) runnerpb.RunnerServer {
	return &srv{
		choreo: choreo,
	}
}

type srv struct {
	runnerpb.UnimplementedRunnerServer
	choreo choreo.Choreo
}

func (r *srv) getBranchContext(branch string) (*choreo.BranchCtx, error) {
	if branch == "" {
		bctx, err := r.choreo.GetBranchStore().GetCheckedOut()
		if bctx == nil {
			return nil, status.Errorf(codes.NotFound, "no checkedout branch found %v", err)
		}
		return bctx, nil
	}
	bctx, err := r.choreo.GetBranchStore().GetStore().Get(store.ToKey(branch))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "err: %s", err.Error())
	}
	return bctx, nil
}

func (r *srv) Start(ctx context.Context, req *runnerpb.Start_Request) (*runnerpb.Start_Response, error) {
	bctx, err := r.getBranchContext("")
	if err != nil {
		return &runnerpb.Start_Response{}, err
	}

	r.choreo.Runner().Start(ctx, bctx)
	return &runnerpb.Start_Response{}, nil
}

func (r *srv) Stop(ctx context.Context, req *runnerpb.Stop_Request) (*runnerpb.Stop_Response, error) {

	r.choreo.Runner().Stop()
	return &runnerpb.Stop_Response{}, nil
}

func (r *srv) Once(req *runnerpb.Once_Request, stream runnerpb.Runner_OnceServer) error {
	ctx := stream.Context()
	bctx, err := r.choreo.GetBranchStore().GetCheckedOut()
	if bctx == nil {
		return status.Errorf(codes.NotFound, "no checkedout branch found %v", err)
	}
	// blocks
	return r.choreo.Runner().RunOnce(ctx, bctx, stream)
}

func (r *srv) Load(ctx context.Context, req *runnerpb.Load_Request) (*runnerpb.Load_Response, error) {
	bctx, err := r.choreo.GetBranchStore().GetCheckedOut()
	if bctx == nil {
		return nil, status.Errorf(codes.NotFound, "no checkedout branch found %v", err)
	}
	if err := r.choreo.Runner().Load(ctx, bctx); err != nil {
		return nil, status.Errorf(codes.Internal, "load data failed %v", err)
	}

	return &runnerpb.Load_Response{}, nil
}

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

/*

import (
	"context"
	"fmt"

	"github.com/henderiw/store"
	recrunner "github.com/kform-dev/choreo/pkg/controller/runner"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func New(choreo choreo.Choreo, runner recrunner.Runner) runnerpb.RunnerServer {
	return &srv{
		choreo: choreo,
		runner: runner,
	}
}

type srv struct {
	runnerpb.UnimplementedRunnerServer
	choreo choreo.Choreo
	runner recrunner.Runner
}

func (r *srv) getBranchContext(branch string) (*choreo.BranchCtx, error) {
	fmt.Println("getBranchContext", branch)
	if branch == "" {
		fmt.Println()
		var bctx *choreo.BranchCtx
		r.choreo.GetBranchStore().GetStore().List(func(k store.Key, bc *choreo.BranchCtx) {
			if bc.State.String() == "CheckedOut" {
				bctx = bc
			}
		})
		if bctx == nil {
			return nil, status.Errorf(codes.NotFound, "no checkedout branch found")
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
	bctx, err := r.getBranchContext(req.Branch)
	if err != nil {
		return nil, err
	}

	if err := r.runner.Start(ctx, bctx); err != nil {
		return &runnerpb.Start_Response{}, status.Errorf(codes.InvalidArgument, "cannot start runner on a branch %s err: %s", req.Branch, err.Error())
	}
	return &runnerpb.Start_Response{}, nil
}

func (r *srv) Stop(ctx context.Context, req *runnerpb.Stop_Request) (*runnerpb.Stop_Response, error) {
	if _, err := r.getBranchContext(req.Branch); err != nil {
		return nil, err
	}
	r.runner.Stop()
	return &runnerpb.Stop_Response{}, nil
}

func (r *srv) Once(ctx context.Context, req *runnerpb.Once_Request) (*runnerpb.Once_Response, error) {
	bctx, err := r.getBranchContext(req.Branch)
	if err != nil {
		return nil, err
	}
	fmt.Println("run once", bctx.Branch)
	rsp, err := r.runner.RunOnce(ctx, bctx.Branch)
	if err != nil {
		return &runnerpb.Once_Response{}, status.Errorf(codes.InvalidArgument, "cannot start runner on a branch %s err: %s", req.Branch, err.Error())
	}
	return rsp, nil
}

func (r *srv) Load(ctx context.Context, req *runnerpb.Load_Request) (*runnerpb.Load_Response, error) {
	bctx, err := r.getBranchContext(req.Branch)
	if err != nil {
		return nil, err
	}
	if err := r.choreo.GetBranchStore().LoadData(ctx, bctx.Branch); err != nil {
		return &runnerpb.Load_Response{}, err
	}
	return &runnerpb.Load_Response{}, nil
}
*/

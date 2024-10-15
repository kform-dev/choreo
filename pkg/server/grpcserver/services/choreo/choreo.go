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

func (r *srv) Commit(ctx context.Context, req *choreopb.Commit_Request) (*choreopb.Commit_Response, error) {
	rsp, err := r.choreo.Get(ctx, &choreopb.Get_Request{})
	if err != nil {
		return &choreopb.Commit_Response{}, err
	}
	if !rsp.Status {
		return &choreopb.Commit_Response{}, status.Error(codes.Unavailable, "choreo not ready to handle request")
	}
	if rsp.ChoreoContext.Production {
		return &choreopb.Commit_Response{}, status.Error(codes.Unavailable, "choreo in production does not allow for commits")
	}
	return r.choreo.GetRootChoreoInstance().CommitWorktree(req.Message)
}

func (r *srv) Push(ctx context.Context, _ *choreopb.Push_Request) (*choreopb.Push_Response, error) {
	rsp, err := r.choreo.Get(ctx, &choreopb.Get_Request{})
	if err != nil {
		return &choreopb.Push_Response{}, err
	}
	if !rsp.Status {
		return &choreopb.Push_Response{}, status.Error(codes.Unavailable, "choreo not ready to handle request")
	}
	if rsp.ChoreoContext.Production {
		return &choreopb.Push_Response{}, status.Error(codes.Unavailable, "choreo in production does not allow for commits")
	}
	return r.choreo.GetRootChoreoInstance().PushBranch(rsp.ChoreoContext.Branch)
}

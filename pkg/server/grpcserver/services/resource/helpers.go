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
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *srv) validatebranch(branch string) error {
	bctx, err := r.choreo.GetBranchStore().GetStore().Get(store.ToKey(branch))
	if err != nil {
		return status.Errorf(codes.NotFound, "branch %s does not exist", branch)
	}
	if bctx.State.String() != "CheckedOut" {
		return status.Errorf(codes.InvalidArgument, "cannot apply to a branch %s which is not checkedout", branch)
	}
	return nil
}

func (r *srv) getCommit(branch string) (*object.Commit, error) {
	bctx, err := r.choreo.GetBranchStore().GetStore().Get(store.ToKey(branch))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "branch %s does not exist", branch)
	}
	return bctx.State.GetCommit(), nil
}

func (r *srv) getStorage(branch string, u *unstructured.Unstructured) (rest.Storage, error) {
	bctx, err := r.choreo.GetBranchStore().GetStore().Get(store.ToKey(branch))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "branch %s does not exist", branch)
	}
	gv, err := schema.ParseGroupVersion(u.GetAPIVersion())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid apiVersion, err: %s", err.Error())
	}
	gvk := gv.WithKind(u.GetKind())

	rctx, err := bctx.APIStore.Get(gvk)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "gvk %s not registered", gvk.String())
	}
	return rctx.Storage, nil
}

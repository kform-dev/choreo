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
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

func (r *srv) getAPIContext(bctx *choreo.BranchCtx, u *unstructured.Unstructured) (*api.ResourceContext, error) {
	gv, err := schema.ParseGroupVersion(u.GetAPIVersion())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid apiVersion, err: %s", err.Error())
	}
	//gvk := gv.WithKind(u.GetKind())
	gk := schema.GroupKind{Group: gv.Group, Kind: u.GetKind()}

	rctx, err := bctx.APIStore.Get(gk)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "gvk %s not registered", gk.String())
	}
	return rctx, nil
}

func convertToInternal(rctx *api.ResourceContext, u *unstructured.Unstructured) {
	if rctx.Internal != nil {
		u.SetAPIVersion(schema.GroupVersion{Group: rctx.Internal.Group, Version: rctx.Internal.Version}.String())
	}
}

func convertFromInternal(rctx *api.ResourceContext, u *unstructured.Unstructured) {
	if rctx.Internal != nil {
		u.SetAPIVersion(schema.GroupVersion{Group: rctx.External.Group, Version: rctx.External.Version}.String())
	}
}

func convertListFromInternal(rctx *api.ResourceContext, ul *unstructured.UnstructuredList) {
	if rctx.Internal != nil {
		ul.SetAPIVersion(schema.GroupVersion{Group: rctx.External.Group, Version: rctx.External.Version}.String())
		for i, item := range ul.Items {
			item.SetAPIVersion(schema.GroupVersion{Group: rctx.External.Group, Version: rctx.External.Version}.String())
			ul.Items[i] = item
		}
	}
}

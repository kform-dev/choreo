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

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
)

var _ State = &NotCheckedOut{}

type NotCheckedOut struct {
	Commit *object.Commit
	Choreo *choreo
	Client resourceclient.Client
}

func (r *NotCheckedOut) String() string { return "NotCheckedOut" }

func (r *NotCheckedOut) Activate(ctx context.Context, branchCtx *BranchCtx) error {
	// the internal apis are already loaded
	// load crds from db in apistore using the apiclient
	loader := &loader.APILoaderCommit2APIStore{
		Client:       r.Choreo.mainChoreoInstance.GetAPIClient(),
		APIStore:     branchCtx.APIStore,
		InternalGVKs: r.Choreo.mainChoreoInstance.GetAPIStore().GetGVKSet(),
		PathInRepo:   r.Choreo.mainChoreoInstance.GetPathInRepo(), // required for the commit read
		DBPath:       r.Choreo.mainChoreoInstance.GetDBPath(),
	}
	if err := loader.Load(ctx, branchCtx.State.GetCommit()); err != nil {
		return err
	}

	// this starts the watchermanager goroutine for the watch to work
	branchCtx.APIStore.Start(ctx)
	return nil
}

func (r *NotCheckedOut) DeActivate(_ context.Context, branchCtx *BranchCtx) error {
	// this stops the watchermanager goroutine
	branchCtx.APIStore.Stop()
	return nil
}

func (r *NotCheckedOut) GetCommit() *object.Commit {
	return r.Commit
}

func (r *NotCheckedOut) LoadData(_ context.Context, branchCtx *BranchCtx) error {
	return fmt.Errorf("loading data in a non checkedout branch %s is not supported", branchCtx.Branch)
}

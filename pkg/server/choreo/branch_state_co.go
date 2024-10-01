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
	"errors"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
)

var _ State = &CheckedOut{}

type CheckedOut struct {
	Choreo               *choreo
	Client               resourceclient.Client
	ChildChoreoInstances []ChoreoInstance
}

func (r *CheckedOut) String() string { return "CheckedOut" }

func (r *CheckedOut) Activate(ctx context.Context, branchCtx *BranchCtx) error {
	// load main choreinst from db/files
	if err := r.loadAPIs(ctx, branchCtx); err != nil {
		return err
	}

	// load child choreoInstances
	if err := r.loadAPIFromUpstreamRefs(ctx, branchCtx); err != nil {
		return err
	}

	if branchCtx.APIStore != nil {
		branchCtx.APIStore.Start(ctx)
	}
	return nil
}

func (r *CheckedOut) loadAPIs(ctx context.Context, branchCtx *BranchCtx) error {
	// load api files to apistore and apiserver
	loader := &loader.APILoaderFile2APIStoreAndAPI{
		Flags:        r.Choreo.flags,
		Client:       r.Client,
		APIStore:     branchCtx.APIStore,
		Branch:       branchCtx.Branch,
		InternalGVKs: r.Choreo.mainChoreoInstance.GetAPIStore().GetGVKSet(),
		RepoPath:     r.Choreo.mainChoreoInstance.GetRepoPath(),
		PathInRepo:   r.Choreo.mainChoreoInstance.GetPathInRepo(),
		DBPath:       r.Choreo.mainChoreoInstance.GetDBPath(),
	}
	if err := loader.Load(ctx); err != nil {
		return err
	}
	return nil
}

func (r *CheckedOut) loadAPIFromUpstreamRefs(ctx context.Context, branchCtx *BranchCtx) error {
	upstreamloader := loader.UpstreamLoader{
		Flags:      r.Choreo.flags,
		Client:     r.Client, // used to upload the upstream ref
		Branch:     branchCtx.Branch,
		RepoPath:   r.Choreo.mainChoreoInstance.GetRepoPath(),
		PathInRepo: r.Choreo.mainChoreoInstance.GetPathInRepo(),
		TempDir:    r.Choreo.mainChoreoInstance.GetTempPath(),
		CallbackFn: r.addChildChoreoInstance,
	}
	if err := upstreamloader.Load(ctx); err != nil {
		return err
	}
	// this is a worktree so we can use it

	var errm error
	for _, childChoreoInstance := range r.ChildChoreoInstances {
		apiStore := api.NewAPIStore()
		loader := &loader.APILoaderCommit2APIStore{
			Client:       childChoreoInstance.GetAPIClient(),
			APIStore:     apiStore,
			InternalGVKs: childChoreoInstance.GetAPIStore().GetGVKSet(),
			PathInRepo:   r.Choreo.mainChoreoInstance.GetPathInRepo(), // required for the commit read
			DBPath:       r.Choreo.mainChoreoInstance.GetDBPath(),
		}
		if err := loader.Load(ctx, childChoreoInstance.GetCommit()); err != nil {
			return err
		}
		// we load the data first to an new apistore
		// after we import to the childresource apistore and the main apistore
		childChoreoInstance.GetAPIStore().Import(apiStore)
		branchCtx.APIStore.Import(apiStore)
	}
	return errm
}

func (r *CheckedOut) addChildChoreoInstance(ctx context.Context, repo repository.Repository, pathInRepo string, flags *genericclioptions.ConfigFlags, commit *object.Commit, annotationValue string) error {
	choreoInstance, err := NewChildChoreoInstance(ctx, repo, pathInRepo, flags, commit, annotationValue)
	if err != nil {
		return err
	}
	r.ChildChoreoInstances = append(r.ChildChoreoInstances, choreoInstance)
	return nil
}

func (r *CheckedOut) GetCommit() *object.Commit {
	return nil
}

func (r *CheckedOut) LoadData(ctx context.Context, branchCtx *BranchCtx) error {
	dataloader := &loader.DataLoader{
		Flags:          r.Choreo.flags,
		Client:         r.Client,
		Branch:         branchCtx.Branch,
		GVKs:           branchCtx.APIStore.GetGVKSet().UnsortedList(),
		RepoPth:        r.Choreo.mainChoreoInstance.GetRepoPath(),
		PathInRepo:     r.Choreo.mainChoreoInstance.GetPathInRepo(),
		APIStore:       branchCtx.APIStore,
		InternalAPISet: r.Choreo.mainChoreoInstance.GetAPIStore().GetGVKSet(),
	}
	if err := dataloader.Load(ctx); err != nil {
		return err
	}

	var errm error
	for _, childChoreoInstance := range r.ChildChoreoInstances {
		loader := &loader.DataLoaderUpstream{
			UpstreamClient:          childChoreoInstance.GetAPIClient(),
			Client:                  r.Client,
			Branch:                  branchCtx.Branch,
			ChildGVKSet:             childChoreoInstance.GetAPIStore().GetGVKSet(),
			UpstreamAnnotationValue: childChoreoInstance.GetAnnotationVal(),
		}
		if err := loader.Load(ctx, childChoreoInstance.GetCommit()); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
	}
	return errm
}

func (r *CheckedOut) DeActivate(_ context.Context, branchCtx *BranchCtx) error {
	// this starts the watchermanager goroutine for the watch to work
	if branchCtx.APIStore != nil {
		branchCtx.APIStore.Stop()
	}
	return nil
}

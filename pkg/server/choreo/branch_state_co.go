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
	// load root choreinst from db/files
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
	rootChoreoInstance := r.Choreo.status.Get().RootChoreoInstance
	loader := &loader.APILoaderFile2APIStoreAndAPI{
		Flags:        r.Choreo.flags,
		Client:       r.Client,
		APIStore:     branchCtx.APIStore,
		Branch:       branchCtx.Branch,
		InternalGVKs: rootChoreoInstance.GetAPIStore().GetGVKSet(),
		RepoPath:     rootChoreoInstance.GetRepoPath(),
		PathInRepo:   rootChoreoInstance.GetPathInRepo(),
		DBPath:       rootChoreoInstance.GetDBPath(),
	}
	if err := loader.Load(ctx); err != nil {
		return err
	}
	return nil
}

func (r *CheckedOut) loadAPIFromUpstreamRefs(ctx context.Context, branchCtx *BranchCtx) error {
	rootChoreoInstance := r.Choreo.status.Get().RootChoreoInstance
	upstreamloader := loader.UpstreamLoader{
		Flags:      r.Choreo.flags,
		Client:     r.Client, // used to upload the upstream ref
		Branch:     branchCtx.Branch,
		RepoPath:   rootChoreoInstance.GetRepoPath(),
		PathInRepo: rootChoreoInstance.GetPathInRepo(),
		TempDir:    rootChoreoInstance.GetTempPath(),
		CallbackFn: r.addChildChoreoInstance,
	}
	if err := upstreamloader.Load(ctx); err != nil {
		return err
	}
	// this is a worktree so we can use it

	var errm error
	for _, childChoreoInstance := range r.ChildChoreoInstances {
		apiStore := api.NewAPIStore()
		loader := &loader.APILoaderFile2APIStoreAndAPI{
			Flags:        r.Choreo.flags,
			Client:       childChoreoInstance.GetAPIClient(),
			APIStore:     apiStore,
			InternalGVKs: childChoreoInstance.GetAPIStore().GetGVKSet(),
			PathInRepo:   childChoreoInstance.GetPathInRepo(), // required for the commit read
			DBPath:       rootChoreoInstance.GetDBPath(),
		}
		if err := loader.LoadFromCommit(ctx, childChoreoInstance.GetCommit()); err != nil {
			return err
		}
		// we load the data first to an new apistore
		// after we import to the childresource apistore and the root apistore
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
	rootChoreoInstance := r.Choreo.status.Get().RootChoreoInstance
	dataloader := &loader.DataLoader{
		Flags:          r.Choreo.flags,
		Client:         r.Client,
		Branch:         branchCtx.Branch,
		GVKs:           branchCtx.APIStore.GetGVKSet().UnsortedList(),
		RepoPth:        rootChoreoInstance.GetRepoPath(),
		PathInRepo:     rootChoreoInstance.GetPathInRepo(),
		APIStore:       branchCtx.APIStore,
		InternalAPISet: rootChoreoInstance.GetAPIStore().GetGVKSet(),
	}
	if err := dataloader.Load(ctx); err != nil {
		return err
	}

	var errm error
	for _, childChoreoInstance := range r.ChildChoreoInstances {
		loader := &loader.DataLoaderUpstream{
			//UpstreamClient:          childChoreoInstance.GetAPIClient(),
			Flags:                   r.Choreo.flags,
			PathInRepo:              childChoreoInstance.GetPathInRepo(),
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

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
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewChildChoreoInstance(ctx context.Context, repo repository.Repository, pathInRepo string, flags *genericclioptions.ConfigFlags, commit *object.Commit, annotationVal string) (ChoreoInstance, error) {
	r := &ChildChoreoInstance{
		repo:          repo,
		pathInRepo:    pathInRepo,
		flags:         flags,
		commit:        commit,
		annotationVal: annotationVal,
	}

	r.apiStoreInternal = api.NewAPIStore()
	if err := r.LoadInternalAPIs(); err != nil {
		return r, err
	}
	r.apiclient = resourceclient.NewAPIStorageClient(r.apiStoreInternal)
	return r, nil

}

type ChildChoreoInstance struct {
	flags         *genericclioptions.ConfigFlags
	repo          repository.Repository
	pathInRepo    string
	commit        *object.Commit
	annotationVal string
	apiclient     resourceclient.Client // apiclient is the client which allows to get access to the local db -> used for commit based api loading

	apiStoreInternal *api.APIStore // this provides the storage layer - w/o the branch view
}

func (r *ChildChoreoInstance) LoadInternalAPIs() error {
	r.apiStoreInternal = api.NewAPIStore()
	loader := loader.APILoaderInternal{
		APIStore:   r.apiStoreInternal,
		Flags:      r.flags,
		DBPath:     r.GetDBPath(),
		PathInRepo: r.GetPathInRepo(),
	}
	return loader.Load(context.Background())
}

func (r *ChildChoreoInstance) GetRepo() repository.Repository {
	return r.repo
}

func (r *ChildChoreoInstance) GetName() string {
	return filepath.Base(r.GetPath())
}

func (r *ChildChoreoInstance) GetPath() string {
	return filepath.Join(r.repo.GetPath(), r.pathInRepo)
}

func (r *ChildChoreoInstance) GetRepoPath() string {
	return r.repo.GetPath()
}

func (r *ChildChoreoInstance) GetPathInRepo() string {
	return r.pathInRepo
}

func (r *ChildChoreoInstance) GetTempPath() string {
	return ""
}

func (r *ChildChoreoInstance) GetDBPath() string {
	return filepath.Join(r.repo.GetPath(), r.pathInRepo, *r.flags.DBPath)
}

func (r *ChildChoreoInstance) GetFlags() *genericclioptions.ConfigFlags {
	return r.flags
}

func (r *ChildChoreoInstance) GetAPIStore() *api.APIStore {
	return r.apiStoreInternal
}

func (r *ChildChoreoInstance) GetCommit() *object.Commit { return r.commit }

func (r *ChildChoreoInstance) GetAPIClient() resourceclient.Client { return r.apiclient }

func (r *ChildChoreoInstance) GetAnnotationVal() string { return r.annotationVal }

func (r *ChildChoreoInstance) Destroy() error {
	return nil
}

func (r *ChildChoreoInstance) CommitWorktree(msg string) (*choreopb.Commit_Response, error) {
	return &choreopb.Commit_Response{}, status.Errorf(codes.Unimplemented, "commitWorktree not implemented on child choreo instance")
}

func (r *ChildChoreoInstance) PushBranch(branch string) (*choreopb.Push_Response, error) {
	return &choreopb.Push_Response{}, status.Errorf(codes.Unimplemented, "commitWorktree not implemented on child choreo instance")
}

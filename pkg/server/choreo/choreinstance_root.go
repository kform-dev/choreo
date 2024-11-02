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
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/repository/git"
	"github.com/kform-dev/choreo/pkg/repository/repogit"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const DummyBranch = "dummy"

type Config struct {
	Path       string
	Repo       repository.Repository
	Commit     *object.Commit
	PathInRepo string
	Cfg        *genericclioptions.ChoreoConfig
}

func NewRootChoreoInstance(ctx context.Context, config *Config) (ChoreoInstance, error) {
	r := &RootChoreoInstance{
		cfg: config.Cfg,
	}

	if config.Repo == nil {
		var err error
		config.Repo, config.PathInRepo, err = getRepoFromPath(ctx, config.Path)
		if err != nil {
			return nil, err
		}
	}
	r.repo = config.Repo
	r.pathInRepo = config.PathInRepo
	r.commit = config.Commit

	// a tempdir is use for the mean choreo instance to load refs
	// (upstream/choreo refs)
	r.tempPath = filepath.Join(r.GetPath(), ".choreo")
	if err := EnsureDir(r.tempPath); err != nil {
		return r, err
	}

	/*
		r.tempPath, err = os.MkdirTemp("choreo", filepath.Base(path))
		if err != nil {
			return r, err
		}
	*/

	r.apiStoreInternal = api.NewAPIStore()
	if err := r.LoadInternalAPIs(); err != nil {
		return r, err
	}
	r.apiclient = resourceclient.NewAPIStorageClient(r.apiStoreInternal)

	return r, nil
}

func getRepoFromPath(ctx context.Context, path string) (repository.Repository, string, error) {
	pathInRepo := "."
	repoPath, git := git.IsPartOfGitRepo(path)
	if !git {
		return nil, "", fmt.Errorf("non git repos not supported yet")
	}
	if path != repoPath {
		// the path is relative within the repo
		pathInRepo = strings.TrimPrefix(path, repoPath+"/")
	}
	repo, err := repogit.NewLocalRepo(ctx, repoPath)
	if err != nil {
		return nil, "", err
	}
	return repo, pathInRepo, nil
}

type RootChoreoInstance struct {
	cfg        *genericclioptions.ChoreoConfig
	repo       repository.Repository
	pathInRepo string
	commit     *object.Commit
	tempPath   string
	apiclient  resourceclient.Client // apiclient is the client which allows to get access to the local db -> used for commit based api loading

	apiStoreInternal *api.APIStore // this provides the storage layer - w/o the branch view
}

func (r *RootChoreoInstance) Destroy() error {
	log := log.FromContext(context.Background())
	if err := r.repo.StashBranch(DummyBranch); err != nil {
		log.Error("stash branch failed", "err", err)
		return err
	}
	if err := r.repo.DeleteBranch(DummyBranch); err != nil {
		log.Error("delete branch failed", "err", err)
		return err
	}
	return nil
}

func (r *RootChoreoInstance) LoadInternalAPIs() error {
	loader := loader.APILoaderInternal{
		APIStore:   r.apiStoreInternal,
		Cfg:        r.cfg,
		DBPath:     r.GetDBPath(),
		PathInRepo: r.GetPathInRepo(),
	}
	return loader.Load(context.Background())
}

func (r *RootChoreoInstance) GetRepo() repository.Repository {
	return r.repo
}

func (r *RootChoreoInstance) GetName() string {
	return filepath.Base(r.GetPath())
}

func (r *RootChoreoInstance) GetPath() string {
	return filepath.Join(r.repo.GetPath(), r.pathInRepo)
}

func (r *RootChoreoInstance) GetRepoPath() string {
	return r.repo.GetPath()
}

func (r *RootChoreoInstance) GetPathInRepo() string {
	return r.pathInRepo
}

func (r *RootChoreoInstance) GetTempPath() string {
	return r.tempPath
}

func (r *RootChoreoInstance) GetDBPath() string {
	return filepath.Join(r.repo.GetPath(), r.pathInRepo, *r.cfg.ServerFlags.DBPath)
}

func (r *RootChoreoInstance) GetConfig() *genericclioptions.ChoreoConfig {
	return r.cfg
}

func (r *RootChoreoInstance) GetAPIStore() *api.APIStore {
	return r.apiStoreInternal
}

func (r *RootChoreoInstance) GetCommit() *object.Commit { return nil }

func (r *RootChoreoInstance) GetAPIClient() resourceclient.Client { return r.apiclient }

func (r *RootChoreoInstance) GetAnnotationVal() string { return "" }

func (r *RootChoreoInstance) CommitWorktree(msg string) (*choreopb.Commit_Response, error) {
	msg, err := r.repo.CommitWorktree(msg, []string{
		filepath.Join(r.pathInRepo, *r.cfg.ServerFlags.InputPath),
	})
	if err != nil {
		return &choreopb.Commit_Response{}, status.Errorf(codes.Internal, "failed pushing branch: %v", err)
	}
	return &choreopb.Commit_Response{
		Message: msg,
	}, nil
}

func (r *RootChoreoInstance) PushBranch(branch string) (*choreopb.Push_Response, error) {
	err := r.repo.PushBranch(branch)
	if err != nil {
		return &choreopb.Push_Response{}, status.Errorf(codes.Internal, "failed pushing branch: %v", err)
	}
	return &choreopb.Push_Response{}, nil
}

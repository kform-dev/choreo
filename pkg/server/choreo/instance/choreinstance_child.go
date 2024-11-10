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

package instance

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/object"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/sdcio/config-diff/schemaloader"
	schemastore "github.com/sdcio/schema-server/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewChildChoreoInstance(ctx context.Context, repo repository.Repository, upstreamRef *choreov1alpha1.UpstreamRef, cfg *genericclioptions.ChoreoConfig, commit *object.Commit, annotationVal string) (ChoreoInstance, error) {
	r := &ChildChoreoInstance{
		upstreamRef:   upstreamRef,
		repo:          repo,
		pathInRepo:    upstreamRef.GetPathInRepo(),
		cfg:           cfg,
		commit:        commit,
		annotationVal: annotationVal,
		children:      []ChoreoInstance{},
	}

	return r, nil

}

type ChildChoreoInstance struct {
	parentName    string
	cfg           *genericclioptions.ChoreoConfig
	upstreamRef   *choreov1alpha1.UpstreamRef
	repo          repository.Repository
	pathInRepo    string
	commit        *object.Commit
	annotationVal string
	children      []ChoreoInstance
	libraries     []*choreov1alpha1.Library
	reconcilers   []*choreov1alpha1.Reconciler
	apiStore      *api.APIStore
}

func (r *ChildChoreoInstance) AddChildChoreoInstance(newchildchoreoinstance ChoreoInstance) error {
	// check if upstream ref is present
	newUpstreamRef := newchildchoreoinstance.GetUpstreamRef()

	if r.upstreamRef.Spec.Type == choreov1alpha1.UpstreamRefType_CRD {
		return fmt.Errorf("cannot load child choreo instances to a crd based choreo instance")
	} else {
		if newUpstreamRef.Spec.Type == choreov1alpha1.UpstreamRefType_All {
			return fmt.Errorf("cannot nest child choreo instances")
		}
	}

	for _, childchoreoinstance := range r.children {
		oldUpstreamRef := childchoreoinstance.GetUpstreamRef()
		if newUpstreamRef.Spec.URL == oldUpstreamRef.Spec.URL &&
			newUpstreamRef.Spec.Directory == oldUpstreamRef.Spec.Directory {
			return fmt.Errorf("conflicting upstreamrefs %s and %s", newUpstreamRef.Name, oldUpstreamRef.Name)
		}
	}
	r.children = append(r.children, newchildchoreoinstance)
	return nil
}

func (r *ChildChoreoInstance) GetUpstreamRef() *choreov1alpha1.UpstreamRef {
	return r.upstreamRef
}

func (r *ChildChoreoInstance) GetChildren() []ChoreoInstance {
	return r.children
}

func (r *ChildChoreoInstance) GetChoreoInstanceName() string {
	return fmt.Sprintf("%s.%s", r.parentName, r.upstreamRef.GetName())
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
	return filepath.Join(r.repo.GetPath(), r.pathInRepo, *r.cfg.ServerFlags.DBPath)
}

func (r *ChildChoreoInstance) GetConfig() *genericclioptions.ChoreoConfig {
	return r.cfg
}

func (r *ChildChoreoInstance) GetInternalAPIStore() *api.APIStore { return nil }

func (r *ChildChoreoInstance) GetCommit() *object.Commit { return r.commit }

func (r *ChildChoreoInstance) GetAPIClient() resourceclient.Client { return nil }

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

func (r *ChildChoreoInstance) IsRootInstance() bool {
	return r.upstreamRef.Spec.Type == choreov1alpha1.UpstreamRefType_All
}

func (r *ChildChoreoInstance) InitAPIs()              { r.apiStore = api.NewAPIStore() }
func (r *ChildChoreoInstance) GetAPIs() *api.APIStore { return r.apiStore }
func (r *ChildChoreoInstance) AddAPIs(apiStore *api.APIStore) {
	if r.apiStore == nil {
		r.apiStore = api.NewAPIStore()
	}
	r.apiStore.Import(apiStore)
}

func (r *ChildChoreoInstance) InitLibraries()                          { r.libraries = []*choreov1alpha1.Library{} }
func (r *ChildChoreoInstance) GetLibraries() []*choreov1alpha1.Library { return r.libraries }
func (r *ChildChoreoInstance) AddLibraries(libraries ...*choreov1alpha1.Library) {
	if r.libraries == nil {
		r.libraries = []*choreov1alpha1.Library{}
	}
	r.libraries = append(r.libraries, libraries...)
}
func (r *ChildChoreoInstance) InitReconcilers()                             { r.reconcilers = []*choreov1alpha1.Reconciler{} }
func (r *ChildChoreoInstance) GetReconcilers() []*choreov1alpha1.Reconciler { return r.reconcilers }
func (r *ChildChoreoInstance) AddReconcilers(reconcilers ...*choreov1alpha1.Reconciler) {
	if r.reconcilers == nil {
		r.reconcilers = []*choreov1alpha1.Reconciler{}
	}
	r.reconcilers = append(r.reconcilers, reconcilers...)
}

func (r *ChildChoreoInstance) InitChildren() {
	r.children = []ChoreoInstance{}
}

func (r *ChildChoreoInstance) SchemaStore() schemastore.Store           { return nil }
func (r *ChildChoreoInstance) SchemaLoader() *schemaloader.SchemaLoader { return nil }

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
	"github.com/go-git/go-git/v5/plumbing/object"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/server/api"
)

type ChoreoInstance interface {
	GetChoreoInstanceName() string
	InitChildren()
	AddChildChoreoInstance(ChoreoInstance) error
	GetChildren() []ChoreoInstance
	GetName() string
	GetRepo() repository.Repository
	GetRepoPath() string
	GetPath() string
	GetTempPath() string
	GetPathInRepo() string
	GetDBPath() string
	GetConfig() *genericclioptions.ChoreoConfig
	GetInternalAPIStore() *api.APIStore // provides the internal apistore, only relevant for rootInstance
	GetCommit() *object.Commit
	GetAPIClient() resourceclient.Client
	GetAnnotationVal() string
	Destroy() error
	CommitWorktree(msg string) (*choreopb.Commit_Response, error)
	PushBranch(branch string) (*choreopb.Push_Response, error)
	GetUpstreamRef() *choreov1alpha1.UpstreamRef
	IsRootInstance() bool

	InitAPIs()
	GetAPIs() *api.APIStore
	AddAPIS(*api.APIStore)
	InitLibraries()
	GetLibraries() []*choreov1alpha1.Library
	AddLibraries(...*choreov1alpha1.Library)
	InitReconcilers()
	GetReconcilers() []*choreov1alpha1.Reconciler
	AddReconcilers(...*choreov1alpha1.Reconciler)
}

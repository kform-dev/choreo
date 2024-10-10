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
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/server/api"
)

type ChoreoInstance interface {
	GetName() string
	GetRepo() repository.Repository
	GetRepoPath() string
	GetPath() string
	GetTempPath() string
	GetPathInRepo() string
	GetDBPath() string
	GetFlags() *genericclioptions.ConfigFlags
	GetAPIStore() *api.APIStore // provides the internal apistore
	GetCommit() *object.Commit
	GetAPIClient() resourceclient.Client
	GetAnnotationVal() string
	Destroy() error
}

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

package repofile

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/kform-dev/choreo/pkg/repository"
)

func New(repopath string, flags *genericclioptions.ConfigFlags) repository.Repository {
	return &repo{
		repopath: repopath,
	}
}

type repo struct {
	repopath string
}

func (r *repo) GetPath() string {
	return r.repopath
}

func (r *repo) BranchExists(branchName string) bool { return branchName == "main" }

func (r *repo) IsBranchCheckedout(branchName string) bool { return true }

func (r *repo) GetBranchCommit(branchName string) (*object.Commit, error) {
	return nil, fmt.Errorf("no commit object in a non git repo")
}

func (r *repo) GetBranchLog(branchName string) ([]*branchpb.Get_Log, error) {
	return nil, fmt.Errorf("not supported on file repo")
}

// GetBranchSet returns a BranchSet with a map of branches
// for easy lookup
func (r *repo) GetBranchSet() repository.BranchSet {
	branchSet := repository.BranchSet{}
	for _, branchObj := range r.GetBranches() {
		branchSet[branchObj.Name] = branchObj
	}
	return branchSet
}

func (r *repo) GetBranches() []*branchpb.BranchObject {
	return []*branchpb.BranchObject{
		{Name: "main", CheckedOut: true},
	}
}
func (r *repo) CreateBranch(branchName string) error { return fmt.Errorf("not supported on file repo") }
func (r *repo) DeleteBranch(branchName string) error { return fmt.Errorf("not supported on file repo") }
func (r *repo) DiffBranch(branchName1, branchName2 string) ([]*branchpb.Diff_Diff, error) {
	return nil, fmt.Errorf("not supported on file repo")
}
func (r *repo) MergeBranch(branchName1, branchName2 string) error {
	return fmt.Errorf("not supported on file repo")
}
func (r *repo) StashBranch(branchName string) error { return fmt.Errorf("not supported on file repo") }
func (r *repo) StreamFiles(branchName string, w *repository.FileWriter) error {
	return fmt.Errorf("not supported on file repo")
}
func (r *repo) Checkout(branchName string) error { return fmt.Errorf("not supported on file repo") }

func (r *repo) GetRefCommit(refName string) (*object.Commit, error) {
	return nil, fmt.Errorf("GetRefCommit not supported in filerepo")
}

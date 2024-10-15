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

package repository

import (
	"io"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
)

type Repository interface {
	//Repo actions
	GetPath() string
	BranchExists(branch string) bool
	IsBranchCheckedout(branch string) bool
	GetRefCommit(ref string) (*object.Commit, error)
	GetBranchCommit(branch string) (*object.Commit, error)
	GetBranchLog(branch string) ([]*branchpb.Get_Log, error)
	GetBranchSet() BranchSet
	GetBranches() []*branchpb.BranchObject
	CreateBranch(branch string) error
	DeleteBranch(branch string) error
	DiffBranch(branch1, branch2 string) ([]*branchpb.Diff_Diff, error)
	MergeBranch(branch1, branch2 string) error
	StashBranch(branch string) error
	StreamFiles(branch string, w *FileWriter) error
	Checkout(branch string) error
	CheckoutCommitRef(commitRef, branch string) (*object.Commit, error)
	CheckoutBranchOrCommitRef(branch, commitRef string) (*object.Commit, error)
	CommitWorktree(msg string, paths []string) (string, error)
	PushBranch(branch string) error
}

type FileWriter struct {
	Writer io.Writer
	Stream branchpb.Branch_StreamFilesServer
}

type BranchSet map[string]*branchpb.BranchObject

func (b BranchSet) Has(branch string) bool {
	_, ok := b[branch]
	return ok
}

func (b BranchSet) Get(branch string) (*branchpb.BranchObject, bool) {
	branchObj, ok := b[branch]
	return branchObj, ok
}

func (b BranchSet) Del(branch string) {
	delete(b, branch)
}

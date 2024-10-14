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

package repogit

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/kform-dev/choreo/pkg/repository"
	lgit "github.com/kform-dev/choreo/pkg/repository/git"
)

func NewLocalRepo(ctx context.Context, repopath string) (repository.Repository, error) {
	gitrepo, err := git.PlainOpen(repopath)
	if err != nil {
		return nil, err
	}
	return &repo{
		repopath: repopath,
		repo:     gitrepo,
	}, nil
}

func NewUpstreamRepo(ctx context.Context, repopath, url, commitHash string) (repository.Repository, *object.Commit, error) {
	gitrepo, commit, err := lgit.Open(ctx, repopath, url, commitHash)
	if err != nil {
		return nil, nil, err
	}

	return &repo{
		repopath: repopath,
		repo:     gitrepo,
	}, commit, nil
}

type repo struct {
	repo     *git.Repository
	repopath string
}

func (r *repo) GetPath() string {
	return r.repopath
}

func (r *repo) BranchExists(branch string) bool {
	if _, err := r.repo.Reference(lgit.BranchName(branch).BranchInLocal(), false); err != nil {
		return r.getBranchFromHeadFIle() == branch
	}
	return true
}

func (r *repo) IsBranchCheckedout(branchName string) bool {
	return branchName == r.getCheckoutBranch()
}

func (r *repo) GetBranchCommit(branchName string) (*object.Commit, error) {
	ref, err := r.repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		return nil, err
	}

	// Get the commit from the reference
	return r.repo.CommitObject(ref.Hash())
}

func (r *repo) GetRefCommit(refName string) (*object.Commit, error) {
	commit, err := lgit.ResolveToCommit(r.repo, refName)
	if err != nil {
		return nil, fmt.Errorf("cannot get commit for ref %s", refName)
	}
	return commit, nil
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
	ctx := context.Background()
	log := log.FromContext(ctx)
	branches := []*branchpb.BranchObject{}
	refIter, err := r.repo.References()
	if err != nil {
		log.Error("failed to get references", "error", err)
		return branches
	}
	err = refIter.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasPrefix(string(ref.Name()), lgit.BranchPrefixInLocalRepo) {
			checkout := false
			if r.getCheckoutBranch() == ref.Name().Short() {
				checkout = true
			}
			branches = append(branches, &branchpb.BranchObject{
				Name:       ref.Name().Short(),
				CheckedOut: checkout,
			})
		}
		return nil
	})
	if err != nil {
		log.Error("failed to iterate over branches", "error", err)
	}
	headFilebranch := r.getBranchFromHeadFIle()
	found := false
	for _, branchObj := range branches {
		if branchObj.Name == headFilebranch {
			found = true
			break
		}
	}
	if !found {
		branches = append(branches, &branchpb.BranchObject{
			Name:       headFilebranch,
			CheckedOut: true,
		})
	}
	return branches
}

func (r *repo) GetBranchLog(branchName string) ([]*branchpb.Get_Log, error) {
	ref, err := r.repo.Reference(lgit.BranchName(branchName).BranchInLocal(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference for branch %s: %s", branchName, err)
	}

	// Get the commit iterator
	iter, err := r.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not get log: %v", err)
	}
	defer iter.Close()

	logs := []*branchpb.Get_Log{}
	// Iterate over the commits
	err = iter.ForEach(func(c *object.Commit) error {
		logs = append(logs, &branchpb.Get_Log{
			CommitHash:  c.Hash.String(),
			AuthorName:  c.Author.Name,
			AuthorEmail: c.Author.Email,
			Date:        c.Author.When.String(),
			Message:     c.Message,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating commits: %v", err)
	}
	return logs, nil
}

func (r *repo) CreateBranch(branchName string) error {
	if r.branchExists(branchName) {
		return nil
	}

	headRef, err := r.repo.Head()
	if err != nil {
		return err
	}
	branchRef := plumbing.NewHashReference(lgit.BranchName(branchName).BranchInLocal(), headRef.Hash())

	err = r.repo.Storer.SetReference(branchRef)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) DeleteBranch(branchName string) error {
	if !r.branchExists(branchName) {
		return nil
	}

	if err := r.Checkout(lgit.MainBranch.BranchInLocal().Short()); err != nil {
		return err
	}

	branchRef := lgit.BranchName(branchName).BranchInLocal()

	// Deleting the branch
	if err := r.repo.Storer.RemoveReference(branchRef); err != nil {
		return fmt.Errorf("failed to delete local branch: %s", err)
	}
	return nil
}

func (r *repo) DiffBranch(branchName1, branchName2 string) ([]*branchpb.Diff_Diff, error) {
	// Get the commit for the first branch.
	commit1, err := r.getBranchCommit(branchName1)
	if err != nil {
		return nil, err
	}

	// Get the commit for the second branch.
	commit2, err := r.getBranchCommit(branchName2)
	if err != nil {
		return nil, err
	}

	// Get the tree for the first commit.
	tree1, err := commit1.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for commit1: %s", err)
	}

	// Get the tree for the second commit.
	tree2, err := commit2.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for commit2: %s", err)
	}

	/*
		// Diff the trees.
		diff, err := object.DiffTree(tree1, tree2)
		if err != nil {
			return fmt.Errorf("failed to diff trees: %s", err)
		}

		for _, change := range diff {
			patch, err := change.Patch()
			if err != nil {
				return fmt.Errorf("failed to get patch: %s", err)
			}
			fmt.Println(patch)
		}
	*/

	// Diff the trees
	changes, err := object.DiffTree(tree1, tree2)
	if err != nil {
		return nil, err
	}

	// Iterate through the changes to print the names of files that have differences
	var errm error
	diffFiles := make([]*branchpb.Diff_Diff, 0, len(changes))
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			errm = errors.Join(errm, err)
			continue
		}

		diffFiles = append(diffFiles, &branchpb.Diff_Diff{
			SrcFileName: change.From.Name,
			DstFileName: change.To.Name,
			Action:      merkletrieAction2branchpbDiffAction(action),
		})
	}

	return diffFiles, errm
}

func merkletrieAction2branchpbDiffAction(action merkletrie.Action) branchpb.Diff_FileAction {
	switch action {
	case merkletrie.Insert:
		return branchpb.Diff_ADDED
	case merkletrie.Delete:
		return branchpb.Diff_DELETED
	case merkletrie.Modify:
		return branchpb.Diff_MODIFIED
	default:
		panic(fmt.Sprintf("unsupported action: %d", action))
	}
}

func (r *repo) MergeBranch(branchName1, branchName2 string) error {
	return nil
}

func (r *repo) StashBranch(branchName string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get workTree %v", err.Error())
	}
	err = w.Reset(&git.ResetOptions{Mode: git.HardReset})
	if err != nil {
		return fmt.Errorf("failed to reset worktree: %v", err)
	}
	return nil
}

func (r *repo) CheckoutCommit(commit *object.Commit, branch string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get workTree %v", err.Error())
	}

	if r.branchExists(branch) {
		if err := r.StashBranch(branch); err != nil {
			return err
		}
		if err := r.DeleteBranch(branch); err != nil {
			return err
		}
	}

	// Checkout the specific commit
	err = w.Checkout(&git.CheckoutOptions{
		Hash:   commit.Hash,
		Branch: lgit.BranchName(branch).BranchInLocal(),
		Create: true,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to create and checkout new branch %s at commit %s: %v", branch, commit.Hash.String(), err)
	}
	branchRef := plumbing.NewHashReference(lgit.BranchName(branch).BranchInLocal(), commit.Hash)

	err = r.repo.Storer.SetReference(branchRef)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) StreamFiles(branchName string, w *repository.FileWriter) error {
	ref, err := r.repo.Reference(lgit.BranchName(branchName).BranchInLocal(), false)
	if err != nil {
		return fmt.Errorf("failed to get reference for branch %s: %s", branchName, err)
	}

	// Get the commit object from the reference
	commit, err := r.repo.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("could not get commit from reference: %v", err)
	}

	// Get the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("could not get tree from commit: %v", err)
	}
	// Walk the tree
	return tree.Files().ForEach(func(f *object.File) error {
		reader, err := f.Reader()
		if err != nil {
			return err
		}
		defer reader.Close()

		content, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		// Stream the content
		if w.Stream != nil {
			// Send file content via gRPC stream
			err = w.Stream.Send(&branchpb.Get_File{
				Name: f.Name,
				Data: string(content),
			})
			if err != nil {
				return err
			}
		} else {
			if _, err = io.Copy(w.Writer, reader); err != nil {
				return err
			}
		}
		return nil

	})
}

func (r *repo) WriteFiles(branchName string, files map[string]string) error {
	// Get the worktree associated with the repository
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %s", err.Error())
	}

	// Check if the branch exists and check it out; create if it does not exist
	err = w.Checkout(&git.CheckoutOptions{
		Branch: lgit.BranchName(branchName).BranchInLocal(),
		Create: false, // Create the branch if it doesn't exist
		Force:  false, // Discard changes in the worktree
		Keep:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %s", err.Error())
	}
	var errm error
	for filePath, content := range files {
		err = os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			errm = errors.Join(errm, fmt.Errorf("failed to write file %s: %s", filePath, err.Error()))
		}
	}
	return errm
}

func (r *repo) Checkout(branchName string) error {
	// Get the worktree associated with the repository
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %s", err.Error())
	}
	// Check if the branch exists and check it out; create if it does not exist
	err = w.Checkout(&git.CheckoutOptions{
		Branch: lgit.BranchName(branchName).BranchInLocal(),
		Create: false, // Create the branch if it doesn't exist
		Force:  false, // Discard changes in the worktree
		Keep:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %s", err.Error())
	}
	return nil
}

func (r *repo) Commit(branchName string, msg string) (string, error) {
	// Get the worktree associated with the repository
	w, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %s", err.Error())
	}

	// Check if the branch exists and check it out; create if it does not exist
	err = w.Checkout(&git.CheckoutOptions{
		Branch: lgit.BranchName(branchName).BranchInLocal(),
		Create: false, // Create the branch if it doesn't exist
		Force:  false, // Discard changes in the worktree
		Keep:   true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to checkout branch: %s", err.Error())
	}
	// Assuming you've made changes in the worktree, stage all changes
	_, err = w.Add(".")
	if err != nil {
		return "", fmt.Errorf("failed to add changes to staging: %s", err.Error())
	}

	// Commit the changes
	commit, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Your Name",
			Email: "your.email@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit changes: %s", err.Error())
	}
	return commit.String(), nil
}

func (r *repo) branchExists(branchName string) bool {
	ref, err := r.repo.Reference(lgit.BranchName(branchName).BranchInLocal(), false)
	if err != nil {
		// this might be an emty
		return false
	}
	return ref != nil
}

func (r *repo) getBranchCommit(branchName string) (*object.Commit, error) {
	ref, err := r.repo.Reference(lgit.BranchName(branchName).BranchInLocal(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference for branch %s: %s", branchName, err)
	}
	commit, err := r.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit for branch %s: %s", branchName, err)
	}
	return commit, nil
}

func (r *repo) getCheckoutBranch() string {
	//log := log.FromContext(context.Background())
	ref, err := r.repo.Head()
	if err != nil {
		// this can be a git w.o commits
		return r.getBranchFromHeadFIle()
	}
	return ref.Name().Short()
}

func (r *repo) getBranchFromHeadFIle() string {
	gitHeadPath := filepath.Join(r.repopath, ".git/HEAD")
	file, err := os.Open(gitHeadPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Read the first line from the .git/HEAD file
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		headContent := scanner.Text()

		// Check if it points to a ref (branch)
		if strings.HasPrefix(headContent, "ref: refs/heads/") {
			branch := strings.TrimPrefix(headContent, "ref: refs/heads/")
			return branch
		} else {
			// If not prefixed by "ref:", it could be a detached HEAD
			return ""
		}
	}
	return ""
}

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

package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/henderiw/logger/log"
)

// IsPartOfGitRepo returns the path of the repo and a boolean
// to indicate if the path is part of a git repo
func IsPartOfGitRepo(path string) (string, bool) {
	for {
		if IsGitRepo(path) {
			return path, true
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", false
		}
		path = parent
	}
}

func IsGitRepo(path string) bool {
	if _, err := git.PlainOpen(path); err != nil {
		return false
	}
	return true
}

// Open open the git repo and either clones or fecthes the remote info
func Open(ctx context.Context, path string, url, refName string) (*git.Repository, *object.Commit, error) {
	cleanup := ""
	defer func() {
		if cleanup != "" {
			os.RemoveAll(cleanup)
		}
	}()

	repo, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		// Repository does not exist, perform a clone
		repo, err := cloneAll(ctx, path, url)
		if err != nil {
			return nil, nil, err
		}
		commit, err := ResolveToCommit(repo, refName)
		if err != nil {
			return nil, nil, err
		}
		cleanup = ""
		return repo, commit, nil
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to open existing repo: %v", err)
	}

	// Repository exists, check if the specific commit is available
	// if so return; if not fetch latest updates and check if commit exists.
	commit, err := ResolveToCommit(repo, refName)
	if err != nil {
		// Commit is not present, fetch it
		if err := fetchUpdates(ctx, repo); err != nil {
			return nil, nil, err
		}
		commit, err := ResolveToCommit(repo, refName)
		if err != nil {
			return nil, nil, err
		}
		cleanup = ""
		return repo, commit, nil
	}

	if err := resetToRemoteHead(ctx, repo, MainBranch.BranchInRemote()); err != nil {
		return nil, nil, err
	}

	// Commit is already present
	return repo, commit, nil
}

func cloneAll(ctx context.Context, path, url string) (*git.Repository, error) {
	log := log.FromContext(ctx)
	// Cloning the repository
	co := &git.CloneOptions{
		URL:          url,
		Depth:        0,
		SingleBranch: false,
		Tags:         git.AllTags,
	}

	var repo *git.Repository
	err := doGitWithAuth(ctx, func(auth transport.AuthMethod) error {
		co.Auth = auth
		var err error
		repo, err = git.PlainClone(path, false, co)
		if err != nil {
			log.Error("Failed to clone with url", "url", url, "error", err)
			return fmt.Errorf("cannot clone repo url %s, err: %v", url, err)
		}
		return nil
	})
	return repo, err
}

func fetchUpdates(ctx context.Context, repo *git.Repository) error {
	log := log.FromContext(ctx)
	// Fetch all branches and tags from the remote
	fetchOptions := &git.FetchOptions{
		RemoteName: OriginName,
		RefSpecs: []config.RefSpec{
			"+refs/heads/*:refs/remotes/origin/*",
			"+refs/tags/*:refs/tags/*",
		},
		Tags: git.AllTags,
	}
	err := repo.Fetch(fetchOptions)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Info("Repository already up-to-date")
			return nil
		}
		log.Error("Failed to fetch updates", "error", err)
		return fmt.Errorf("failed to fetch updates: %v", err)
	}
	return nil
}

func resetToRemoteHead(ctx context.Context, repo *git.Repository, branch plumbing.ReferenceName) error {
	log := log.FromContext(ctx)
	ref, err := repo.Reference(branch, true)
	if err != nil {
		log.Error("Failed to get reference", "branch", branch, "error", err)
		return fmt.Errorf("failed to get reference for branch %s: %v", branch, err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		log.Error("Failed to get worktree", "error", err)
		return fmt.Errorf("cannot get worktree: %v", err)
	}
	err = wt.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: ref.Hash(),
	})
	if err != nil {
		log.Error("Failed to reset worktree", "ref", ref.Hash(), "error", err)
		return fmt.Errorf("cannot reset worktree: %v", err)
	}
	return nil
}

// CheckoutRef checks out a specific ref
func CheckoutRef(repo *git.Repository, refName string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("cannot get worktree: %v", err)
	}

	fmt.Println("refName", refName)
	hash, err := GetHash(repo, refName)
	if err != nil {
		return fmt.Errorf("cannot get hash: %v", err)
	}
	fmt.Println("refName", hash.String())
	err = wt.Checkout(&git.CheckoutOptions{
		Hash: *hash,
	})
	if err != nil {
		return fmt.Errorf("cannot checkout commit %s, err: %v", refName, err)
	}
	return nil
}

func CloneNonExisting(ctx context.Context, path, url string) (*git.Repository, error) {
	var err error
	var repo *git.Repository
	// init clone options
	co := &git.CloneOptions{
		Depth: 1,
		URL:   url,
		//ReferenceName: branchRef, // remote reference
		//SingleBranch:  true,
	}

	// perform clone
	err = doGitWithAuth(ctx, func(auth transport.AuthMethod) error {
		co.Auth = auth
		repo, err = git.PlainClone(path, false, co)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot fetch repo, err: %s", err)
	}
	return repo, nil
}

func CloneExisting(_ context.Context, path, _ string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}
	return repo, nil

	/*
		remote, err := repo.Remote(OriginName)
		if err != nil {
			return nil, fmt.Errorf("cannot get remote: %s", err)
		}

		// checking that the configured remote equals the provided remote
		if remote.Config().URLs[0] != url {
			return nil, fmt.Errorf("provided url %q differs from configured url %q", remote.Config().URLs[0], url)
		}

		// We have a shallow clone - we cannot simply pull new changes. See:
		// https://stackoverflow.com/a/41081908 for a detailed explanation of how this works
		// We need to fetch and then reset && clean to update the repo contents - otherwise can be left with
		// 'object not found' error, presumably because there is no link between the two commits (due to shallow clone)
		localRefName := MainBranch.BranchInLocal()
		remoteRefName := MainBranch.BranchInRemote()

		refSpec := config.RefSpec(fmt.Sprintf("+%s:%s", localRefName, remoteRefName))
		err = doGitWithAuth(ctx, func(auth transport.AuthMethod) error {
			return repo.FetchContext(ctx, &git.FetchOptions{
				Depth: 1,
				Auth:  auth,
				Force: true,
				Prune: true,
				RefSpecs: []config.RefSpec{
					refSpec,
				},
			})
		})
		switch {
		case errors.Is(err, git.NoErrAlreadyUpToDate):
			err = nil
		}
		if err != nil {
			return nil, fmt.Errorf("cannot fetch repo, err: %s", err)
		}
		return repo, nil
	*/
}

// doGitWithAuth fetches auth information for git and provides it
// to the provided function which performs the operation against a git repo.
func doGitWithAuth(ctx context.Context, op func(transport.AuthMethod) error) error {
	log := log.FromContext(ctx)
	auth, err := getAuthMethod(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to obtain git credentials: %w", err)
	}
	err = op(auth)
	if err != nil {
		if !errors.Is(err, transport.ErrAuthenticationRequired) {
			return err
		}
		log.Info("Authentication failed. Trying to refresh credentials")
		// TODO: Consider having some kind of backoff here.
		auth, err := getAuthMethod(ctx, true)
		if err != nil {
			return fmt.Errorf("failed to obtain git credentials: %w", err)
		}
		return op(auth)
	}
	return nil
}

// getAuthMethod fetches the credentials for authenticating to git. It caches the
// credentials between calls and refresh credentials when the tokens have expired.
func getAuthMethod(_ context.Context, _ bool) (transport.AuthMethod, error) {
	// If no secret is provided, we try without any auth.
	return nil, nil
}

// resolveToCommit takes a repository and a reference name (tag or commit hash) and resolves it to a commit object
func GetHash(repo *git.Repository, refName string) (*plumbing.Hash, error) {
	// Check if the refName is a hash directly
	hash, err := repo.ResolveRevision(plumbing.Revision(refName))
	if err == nil {
		// It's a direct hash or a lightweight tag
		return hash, nil
	}

	// Resolve reference, could be an annotated tag
	ref, err := repo.Reference(plumbing.ReferenceName(refName), true)
	if err != nil {
		return nil, fmt.Errorf("could not resolve reference: %v", err)
	}

	// Check if the reference is a tag object (annotated tag)
	if ref.Type() == plumbing.HashReference {
		hash := ref.Hash()
		return &hash, nil
	}
	return nil, fmt.Errorf("cannot find hash for refName %s", refName)
}

// resolveToCommit takes a repository and a reference name (tag or commit hash) and resolves it to a commit object
func ResolveToCommit(repo *git.Repository, refName string) (*object.Commit, error) {
	// Check if the refName is a hash directly
	hash, err := repo.ResolveRevision(plumbing.Revision(refName))
	if err == nil {
		// It's a direct hash or a lightweight tag
		return repo.CommitObject(*hash)
	}

	// Resolve reference, could be an annotated tag
	ref, err := repo.Reference(plumbing.ReferenceName(refName), true)
	if err != nil {
		return nil, fmt.Errorf("could not resolve reference: %v", err)
	}

	// Check if the reference is a tag object (annotated tag)
	if ref.Type() == plumbing.HashReference {
		return repo.CommitObject(ref.Hash())
	}

	// If it's an annotated tag, dereference the tag object to get the commit
	tag, err := repo.TagObject(ref.Hash())
	if err != nil {
		// Not a tag object, try to return as a commit directly
		return repo.CommitObject(ref.Hash())
	}

	return tag.Commit()
}

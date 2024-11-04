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
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/henderiw/store/memory"
	"github.com/henderiw/store/watch"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/server/api"
)

func NewBranchStore(choreo *choreo) *BranchStore {
	return &BranchStore{
		choreo: choreo,
		store:  memory.NewStore[*BranchCtx](newBranchCtx),
	}
}

func newBranchCtx() *BranchCtx { return &BranchCtx{} }

type BranchStore struct {
	choreo *choreo
	store  store.Storer[*BranchCtx]
}

type BranchCtx struct {
	State    State
	Branch   string
	APIStore *api.APIStore
}

func (r *BranchStore) Update(ctx context.Context, branches []*branchpb.BranchObject) error {
	var errm error
	for _, branchObj := range branches {
		var newState State
		newState = &CheckedOut{}

		if !branchObj.CheckedOut {
			newState = &NotCheckedOut{}
		}

		if err := r.update(ctx, branchObj.Name, newState); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
	}
	return errm
}

func (r *BranchStore) update(ctx context.Context, branch string, newState State) error {
	log := log.FromContext(ctx)
	key := store.ToKey(branch)
	var oldState State
	branchCtx, err := r.store.Get(key)
	if err == nil {
		// branch exists
		if newState.String() == branchCtx.State.String() {
			// no state transition -> do nothing
			return nil
		}
		oldState = branchCtx.State
	}

	// import the internal apis for storage purpose
	apiStore := api.NewAPIStore()
	apiStore.Import(r.choreo.status.Get().RootChoreoInstance.GetInternalAPIStore())

	newBranchCtx := &BranchCtx{
		State:    newState,
		Branch:   branch,
		APIStore: apiStore,
	}
	if err := r.store.Apply(key, newBranchCtx); err != nil {
		return err
	}

	log.Info("branchstore update", "branch", branch, "state change", fmt.Sprintf("%s->%s", oldState, newState))
	if err := r.handleTransition(ctx, newBranchCtx, oldState, newState); err != nil {
		return err
	}
	// apply the changes to the branchCtx after the state transition
	if err := r.store.Apply(key, newBranchCtx); err != nil {
		return err
	}
	return nil
}

func (r *BranchStore) Delete(ctx context.Context, branchSet repository.BranchSet) error {
	log := log.FromContext(ctx)
	// to be executed before updates, otherwise this iwill not work
	var errm error
	branchesToBeDeleted := []string{}
	r.store.List(func(k store.Key, branchCtx *BranchCtx) {
		branch := k.Name
		if _, exists := branchSet.Get(branch); !exists {
			if err := r.handleTransition(ctx, branchCtx, branchCtx.State, nil); err != nil {
				errm = errors.Join(errm, err)
				return
			}
			branchesToBeDeleted = append(branchesToBeDeleted, branchCtx.Branch)

		}
	})
	// when we do a list in the store and you do another store operation the mutex will lock
	for _, branch := range branchesToBeDeleted {
		log.Info("branchstore delete", "branch", branch)
		if err := r.store.Delete(store.ToKey(branch)); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
		log.Info("branchstore deleted", "branch", branch)
	}
	return errm
}

func (r *BranchStore) handleTransition(ctx context.Context, branchCtx *BranchCtx, oldState, newState State) error {
	if oldState != nil {
		if err := oldState.DeActivate(ctx, branchCtx); err != nil {
			return err
		}
	}
	if newState != nil {
		return newState.Activate(ctx, branchCtx)
	}
	return nil
}

func (r *BranchStore) GetStore() store.Storer[*BranchCtx] {
	return r.store
}

func (r *BranchStore) GetCheckedOut() (*BranchCtx, error) {
	var bctx *BranchCtx
	r.store.List(func(k store.Key, bc *BranchCtx) {
		if bc.State.String() == "CheckedOut" {
			bctx = bc
		}
	})
	if bctx == nil {
		return nil, fmt.Errorf("no checkedout branch found")
	}
	return bctx, nil
}

/*
func (r *BranchStore) Load(ctx context.Context, branch string) error {
	bctx, err := r.store.Get(store.ToKey(branch))
	if err != nil {
		return err
	}
	return bctx.State.Load(ctx, bctx)
}
*/

func (r *BranchStore) WatchBranches(ctx context.Context, opts ...store.ListOption) (watch.WatchInterface[*BranchCtx], error) {
	return r.store.Watch(ctx, opts...)
}

func (r *BranchStore) WatchAPIResources(ctx context.Context, opts ...resourceclient.ListOption) (watch.WatchInterface[*api.ResourceContext], error) {
	o := &resourceclient.ListOptions{}
	o.ApplyOptions(opts)

	branchCtx, err := r.store.Get(store.ToKey(o.Branch))
	if err != nil {
		return nil, err
	}

	return branchCtx.APIStore.Watch(ctx, &store.ListOptions{})
}

func (r *BranchStore) UpdateBranchCtx(branchCtx *BranchCtx) error {
	return r.store.Apply(store.ToKey(branchCtx.Branch), branchCtx)
}

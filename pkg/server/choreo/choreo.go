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
	"os"
	"path/filepath"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
)

type Choreo interface {
	Start(ctx context.Context)
	GetMainChoreoInstance() ChoreoInstance
	GetBranchStore() *BranchStore
}

func New(ctx context.Context, path string, flags *genericclioptions.ConfigFlags) (Choreo, error) {
	mainChoreoInstance, err := NewMainChoreoInstance(ctx, path, flags)
	if err != nil {
		return nil, err
	}
	r := &choreo{
		flags:              flags,
		name:               filepath.Base(path),
		gitrepo:            false,
		path:               path,
		mainChoreoInstance: mainChoreoInstance,
	}
	r.branchStore, err = NewBranchStore(r)
	if err != nil {
		return r, err
	}
	return r, nil
}

type choreo struct {
	flags              *genericclioptions.ConfigFlags
	gitrepo            bool
	name               string
	path               string
	mainChoreoInstance ChoreoInstance

	client      resourceclient.Client
	branchStore *BranchStore
}

func (r *choreo) GetBranchStore() *BranchStore {
	return r.branchStore
}

func (r *choreo) GetMainChoreoInstance() ChoreoInstance {
	return r.mainChoreoInstance
}

func (r *choreo) Start(ctx context.Context) {
	log := log.FromContext(ctx)
	var err error
	r.client, err = r.flags.ToResourceClient()
	if err != nil {
		panic(err)
	}
	r.branchStore.store.Start(ctx)
	defer r.branchStore.store.Stop()

	// Ticker to check the repository every 10 seconds
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	defer os.RemoveAll(r.mainChoreoInstance.GetTempPath())

	if err := r.updateBranches(ctx); err != nil {
		log.Error("update branches failed", "err", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := r.updateBranches(ctx); err != nil {
				log.Error("update branches failed", "err", err)
			}

		case <-ctx.Done():
			time.Sleep(1 * time.Second)
			fmt.Println("choreo done")
			return
		}
	}
}

func (r *choreo) updateBranches(ctx context.Context) error {
	if err := r.branchStore.Delete(ctx, r.mainChoreoInstance.GetRepo().GetBranchSet()); err != nil {
		return err
	}
	if err := r.branchStore.Update(ctx, r.mainChoreoInstance.GetRepo().GetBranches()); err != nil {
		return err
	}
	return nil
}

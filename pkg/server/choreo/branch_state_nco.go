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

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
)

var _ State = &NotCheckedOut{}

type NotCheckedOut struct {
	Commit *object.Commit
	Choreo *choreo
	Client resourceclient.Client
}

func (r *NotCheckedOut) String() string { return "NotCheckedOut" }

func (r *NotCheckedOut) Activate(ctx context.Context, branchCtx *BranchCtx) error {
	// this starts the watchermanager goroutine for the watch to work
	branchCtx.APIStore.Start(ctx)
	return nil
}

func (r *NotCheckedOut) DeActivate(_ context.Context, branchCtx *BranchCtx) error {
	// this stops the watchermanager goroutine
	branchCtx.APIStore.Stop()
	return nil
}

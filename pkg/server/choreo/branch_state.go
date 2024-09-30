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
	"reflect"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type Event int

const (
	Activate Event = iota
	DeActivate
)

func (r Event) String() string {
	switch r {
	case Activate:
		return "Activate"
	case DeActivate:
		return "DeActivate"
	default:
		return reflect.TypeOf(r).Name()
	}
}

type State interface {
	String() string
	Activate(ctx context.Context, branchCtx *BranchCtx) error
	DeActivate(ctx context.Context, branchCtx *BranchCtx) error
	LoadData(ctx context.Context, branchCtx *BranchCtx) error
	GetCommit() *object.Commit
}

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

package view

import (
	"context"
	"sync"

	"github.com/rivo/tview"
)

type Pages struct {
	*tview.Pages

	app       *App
	mainPages map[string]Page
	stack     *stack
}

func NewPages(ctx context.Context) *Pages {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}

	r := &Pages{
		app:       app,
		Pages:     tview.NewPages(),
		mainPages: map[string]Page{},
		stack:     newStack(),
	}
	r.mainPages["dummy"] = NewDummy(ctx, r)
	r.mainPages["resources"] = NewResources(ctx, r)
	r.mainPages["branch"] = NewBranchPage(ctx, r)

	return r
}

func (r *Pages) RegisterPageAction(ctx context.Context) {
	for _, page := range r.mainPages {
		page.RegisterPageAction(ctx)
	}
}

type Page interface {
	//TviewPrimitive() tview.Primitive

	ActivatePage(ctx context.Context)

	// registers the keys of the page to the
	// actions per app and the cmd header
	RegisterPageAction(ctx context.Context)
}

type stack struct {
	stack []Page
	m     sync.RWMutex
}

func newStack() *stack {
	return &stack{
		stack: make([]Page, 0),
	}
}

// Add adds a new key action.
func (r *stack) Next(p Page) {
	r.m.Lock()
	defer r.m.Unlock()

	r.stack = append(r.stack, p)
}

func (r *stack) Swap(p Page) {
	if len(r.stack) == 0 {
		return
	}
	r.stack[len(r.stack)-1] = p
}

// Add adds a new key action.
func (r *stack) Back() {
	r.m.Lock()
	defer r.m.Unlock()

	if len(r.stack) == 0 {
		return
	}
	r.stack = r.stack[:len(r.stack)-1]
}

func (r *stack) Last() Page {
	if len(r.stack) == 0 {
		return nil
	}
	return r.stack[len(r.stack)-1]
}

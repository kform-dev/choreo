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
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/rivo/tview"
)

type App struct {
	factory util.Factory

	*tview.Application

	mainFlex *tview.Flex
	header   *Header
	pages    *Pages // this might have to become a stack
	//views            map[string]tview.Primitive
	ticker           *ticker
	branchSelection  string
	checkedOutBranch string
	actions          *KeyActions
}

func NewApp(f util.Factory) *App {
	app := &App{
		factory:     f,
		Application: tview.NewApplication(),
		actions:     NewKeyActions(),
	}

	return app
}

func (a *App) Init(ctx context.Context) error {
	ctx = context.WithValue(ctx, KeyApp, a)

	// Handle input is handled generically
	a.Application.SetInputCapture(a.HandleInput)

	// initialize the header
	a.header = NewHeader(a)
	// initialize the pages
	a.pages = NewPages(ctx)

	// setup layout
	a.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 6, 1, true). // 4 sets min height to 4
		AddItem(a.pages, 0, 6, true)

	a.SetRoot(a.mainFlex, true)
	// initialize the actions within the context
	a.pages.mainPages["resources"].ActivatePage(ctx)
	return nil
}

// In your page implementation
func (a *App) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		if action, exists := a.actions.Get(string(event.Rune())); exists {
			action.Action()
			return nil // Stop further processing
		}
	}
	return event // default pass-through
}

// Run starts the application loop.
func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	//go a.watchBranch(ctx)

	if err := a.Application.Run(); err != nil {
		return err
	}
	return nil
}

func (a *App) watchBranch(ctx context.Context) {
	rspch := a.factory.GetBranchClient().Watch(ctx, &branchpb.Watch_Request{Options: &branchpb.Watch_Options{}})

	branches := map[string]int{}
	// TODO columnwidth

	for {
		select {
		case <-ctx.Done():
			// watch stoppped
			return
		case event, ok := <-rspch:
			if !ok {
				return
			}
			switch event.EventType {
			case branchpb.Watch_ADDED:

			case branchpb.Watch_MODIFIED:
				branch := event.BranchObj.Name
				if _, ok := branches[branch]; ok {

				}

			case branchpb.Watch_DELETED:
			default:

			}
		}
	}
}

func (a *App) stop(key *tcell.EventKey) *tcell.EventKey {
	if a.ticker != nil {
		a.ticker.stop()
	}
	a.Stop()

	a.actions.List(func(k string, ka KeyAction) {
		fmt.Println("list key", k)
	})
	fmt.Println("key", key.Key())

	switch key.Key() {
	case tcell.KeyRune:
		fmt.Println("rune", key.Rune())

		_, exists := a.actions.Get(string(key.Rune()))
		fmt.Println("key", exists)
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// AsKey converts rune to keyboard key.,.
func AsKey(evt *tcell.EventKey) tcell.Key {
	if evt.Key() != tcell.KeyRune {
		return evt.Key()
	}
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}

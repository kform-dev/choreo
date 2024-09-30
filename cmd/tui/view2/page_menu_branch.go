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
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type BranchMenu struct {
	*tview.Table

	app   *App
	name  string
	title string

	actions  *KeyActions
	keyStyle tcell.Style
	valStyle tcell.Style
}

// NewInfo creates a new Info view.
func NewBranchMenu(app *App) *BranchMenu {

	r := &BranchMenu{
		Table:   tview.NewTable(),
		name:    "branchmenu",
		title:   "BranchMenu",
		app:     app,
		actions: NewKeyActions(),
		keyStyle: tcell.StyleDefault.
			Foreground(tcell.ColorFuchsia).
			Background(tcell.ColorBlack),
		valStyle: tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlack),
	}
	// views are not needed as these pages are not selectable

	return r
}

func (r *BranchMenu) InitPageAction() {
	r.actions = NewKeyActions()
	r.actions.Add(KeyAction{Key: Key0, ShortName: "main", Action: func() { r.app.branchSelection = "main" }})
	r.actions.Add(KeyAction{Key: Key1, ShortName: "checkedout", Action: func() { r.app.branchSelection = "checkedout" }})
}

// ActivateAction
// propogates the actions
// populate the cmdMenu
func (r *BranchMenu) ActivatePageAction() {
	// clear table
	r.Table.Clear()
	// this provides a mapping to rows
	sortedKeys := r.actions.MapKeys()

	r.actions.List(func(k string, ka KeyAction) {
		// propgate the action
		r.app.actions.Add(ka)
		// add the keys to the table
		SetMenuRow(r.Table, sortedKeys[k], k, ka.ShortName, r.keyStyle, r.valStyle)
	})
}

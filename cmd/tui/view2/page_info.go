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

type Info struct {
	*tview.Table

	app   *App
	name  string
	title string
}

// NewInfo creates a new Info view.
func NewInfo(app *App) *Info {
	r := &Info{
		Table: tview.NewTable(),
		name:  "info",
		title: "Info",
		app:   app,
	}
	//r.SetBorder(true)
	//r.SetTitle("info")
	r.Table.SetBackgroundColor(tcell.ColorBlack)

	r.SetKeyValue(0, "chooreo version", "v0.0.0")
	r.SetBackgroundColor(tcell.ColorBlack)
	return r
}

func (r *Info) SetKeyValue(row int, key, value string) {
	r.SetCell(
		row,
		0, // column for key
		tview.NewTableCell(key).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorFuchsia),
	)
	r.SetCell(
		row,
		1, // column for value
		tview.NewTableCell(value).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorFuchsia),
	)
}

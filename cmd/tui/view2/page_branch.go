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
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/rivo/tview"
)

type BranchPage struct {
	HeaderTable *tview.Table
	DataTable   *tview.Table

	app    *App
	name   string
	title  string
	view   *tview.Flex
	style  tcell.Style
	cancel func()
}

func NewBranchPage(ctx context.Context, pages *Pages) Page {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}

	r := &BranchPage{
		name:        "branches",
		title:       "Branches",
		HeaderTable: tview.NewTable(),
		DataTable:   tview.NewTable(),
		app:         app,
		style: tcell.StyleDefault.
			Foreground(tcell.ColorLightSkyBlue).
			Background(tcell.ColorBlack),
	}
	r.SetTable(ctx)
	// TODO Mousecapture
	r.SetView()
	pages.AddPage(r.name, r.view, true, true)
	return r
}

func (r *BranchPage) SetTable(ctx context.Context) {
	r.HeaderTable.SetBackgroundColor(tcell.ColorBlack)
	r.HeaderTable.SetFixed(1, 0) // set header fixed
	r.DataTable.SetBackgroundColor(tcell.ColorBlack)
	r.DataTable.SetFixed(0, 0)
	r.DataTable.
		SetSelectable(true, false).
		Select(0, 0).
		SetDoneFunc(func(key tcell.Key) {
			r.DeActivatePage(ctx)
		}).
		SetSelectedFunc(func(row, column int) {
		}).
		ScrollToBeginning()
}

func (r *BranchPage) SetView() {
	view := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.HeaderTable, 1, 1, false).
		AddItem(r.DataTable, 0, 1, true)
	view.
		SetBorder(true).
		SetTitle(r.title).
		SetBlurFunc(func() {
			if r.app.ticker != nil {
				r.app.ticker.stop()
				r.app.ticker = nil
			}
		}).
		SetBorderStyle(r.style)
	r.view = view
}

func (r *BranchPage) RegisterPageAction(ctx context.Context) {
	ka := KeyAction{
		Key:       KeyB,
		ShortName: r.name,
		Action: func() {
			r.ActivatePage(ctx)
		},
	}
	// propagate keys
	r.app.actions.Add(ka)
	// add key to menu table
	r.app.header.cmdMenu.actions.Add(ka)
}

func (r *BranchPage) DeActivatePage(ctx context.Context) {
	r.StopStream(ctx)
}

func (r *BranchPage) ActivatePage(ctx context.Context) {
	// init action
	r.app.actions = NewKeyActions()
	r.app.header.InitPageAction()
	r.app.pages.RegisterPageAction(ctx) // central main page keys

	// activate page action
	r.app.header.ActivatePageAction("main")

	r.app.pages.SwitchToPage(r.name)
	r.app.SetFocus(r.DataTable)
	r.DataTable.SetInputCapture(r.HandleInput)
	go r.StartStream(ctx)
	//r.app.ticker.start()
	r.app.ForceDraw()
}

// In your page implementation
func (r *BranchPage) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	return event // Propagate to global handler
}

func (r *BranchPage) StopStream(ctx context.Context) {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *BranchPage) StartStream(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	rspch := r.app.factory.GetBranchClient().Watch(ctx, &branchpb.Watch_Request{Options: &branchpb.Watch_Options{}})

	idset := NewIdentifierSet()

	r.app.QueueUpdateDraw(func() {
		r.DataTable.Clear()
		SetTableHeader(r.HeaderTable, []int{10, 20}, "Name", "CheckedOut")
	})

	for {
		select {
		case <-ctx.Done():
			// watch stoppped
			return
		case event, ok := <-rspch:
			if !ok {
				return
			}
			r.app.QueueUpdateDraw(func() {
				switch event.EventType {
				case branchpb.Watch_ADDED, branchpb.Watch_MODIFIED:
					id, exists := idset.AddIdentifier(event.BranchObj.Name)
					if exists {
						UpdateRowAt(r.DataTable, id, nil, GetBranchRowData(event.BranchObj), r.style)
					} else {
						InsertRowAt(r.DataTable, id, nil, GetBranchRowData(event.BranchObj), r.style)
					}

				case branchpb.Watch_DELETED:
					id := idset.DeleteIdentifier(event.BranchObj.Name)
					if id != -1 {
						DeleteRow(r.DataTable, id)
					}
				default:

				}
				columnWidth := CalculateMaxWidths(r.DataTable)
				SetTableHeader(r.HeaderTable, columnWidth, "Name", "CheckedOut")
				r.HeaderTable.ScrollToBeginning()
				r.DataTable.ScrollToBeginning()
				//r.app.ForceDraw()
			})
		}
	}
}

/*
func (r *BranchPage) actvateChildPage(ctx context.Context, row, column int) error {
	resource, ok := r.DataTable.GetCell(row, column).GetReference().(*choreov1alpha1.APIResourceGroup)
	if !ok {
		return fmt.Errorf("%s addChildPage resource not available in row %d, column %d", r.name, row, column)
	}
	p := NewResource(ctx, r, resource)
	p.ActivatePage(ctx)
	return nil
}
*/

func GetBranchRowData(branch *branchpb.BranchObject) []string {
	row := []string{}
	row = append(row, branch.Name)
	row = append(row, fmt.Sprintf("%v", branch.CheckedOut))
	return row
}

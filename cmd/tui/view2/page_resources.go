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
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Resources struct {
	HeaderTable *tview.Table
	DataTable   *tview.Table

	app    *App
	name   string
	title  string
	view   *tview.Flex
	style  tcell.Style
	cancel func()
}

func NewResources(ctx context.Context, pages *Pages) Page {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}

	r := &Resources{
		name:        "resources",
		title:       "Resources",
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

func (r *Resources) SetTable(ctx context.Context) {
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
			r.actvateChildPage(ctx, row, column)
		}).
		ScrollToBeginning()
}

func (r *Resources) SetView() {
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

func (r *Resources) RegisterPageAction(ctx context.Context) {
	ka := KeyAction{
		Key:       KeyR,
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

func (r *Resources) DeActivatePage(ctx context.Context) {
	r.StopStream(ctx)
}

func (r *Resources) ActivatePage(ctx context.Context) {
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
	/*
			r.app.ticker = NewTicker(r.app.Application, 2*time.Second, func() {
				r.Update(ctx)
			})

		r.app.ticker.start()
	*/
	r.app.ForceDraw()
}

// In your page implementation
func (r *Resources) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	return event // Propagate to global handler
}

func (r *Resources) StopStream(ctx context.Context) {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Resources) StartStream(parentCtx context.Context) {
	for {
		ctx, cancel := context.WithCancel(parentCtx)
		r.cancel = cancel

		branch := r.app.factory.GetBranch()
		proxy := r.app.factory.GetProxy()

		rspch := r.app.factory.GetDiscoveryClient().Watch(ctx, &discoverypb.Watch_Request{
			ProxyName:      proxy.Name,
			ProxyNamespace: proxy.Namespace,
			Branch:         branch,
			Options:        &discoverypb.Watch_Options{},
		})

		// Process the stream
		r.processStream(ctx, rspch)

		// Cleanup after stream is closed
		cancel() // Clean up the current stream's context

		// Check if parent context has been cancelled before restarting
		select {
		case <-parentCtx.Done():
			return // Exit the function, no restart
		default:
			fmt.Println("restart")
			time.Sleep(time.Second * 2) // Throttle the restart mechanism
		}
	}
}

func (r *Resources) processStream(ctx context.Context, rspch <-chan *discoverypb.Watch_Response) {
	idset := NewIdentifierSet()

	r.app.QueueUpdateDraw(func() {
		r.DataTable.Clear()
		SetTableHeader(r.HeaderTable, []int{10, 10, 10, 10, 10, 10}, "GroupResource", "Version", "Kind", "ListKind", "Namespaced", "Categories")
	})
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled, likely by parent context or stream closure
			return
		case event, ok := <-rspch:
			if !ok {
				// Stream closed by the server, exit the loop to restart the stream
				return
			}

			// Handle events normally
			r.app.QueueUpdateDraw(func() {
				r.handleEvent(event, idset)
			})
		}
	}
}

func (r *Resources) handleEvent(event *discoverypb.Watch_Response, idset *IdentifierSet) {
	switch event.EventType {
	case discoverypb.Watch_ADDED, discoverypb.Watch_MODIFIED:
		id, exists := idset.AddIdentifier(getResourceID(event.ApiResource))
		if exists {
			UpdateRowAt(r.DataTable, id, event.ApiResource, GetResourceRowData(event.ApiResource), r.style)
		} else {
			InsertRowAt(r.DataTable, id, event.ApiResource, GetResourceRowData(event.ApiResource), r.style)
		}

	case discoverypb.Watch_DELETED:
		id := idset.DeleteIdentifier(getResourceID(event.ApiResource))
		if id != -1 {
			DeleteRow(r.DataTable, id)
		}
	default:

	}
	columnWidth := CalculateMaxWidths(r.DataTable)
	SetTableHeader(r.HeaderTable, columnWidth, "GroupResource", "Version", "Kind", "ListKind", "Namespaced", "Categories")
	r.HeaderTable.ScrollToBeginning()
	r.DataTable.ScrollToBeginning()
}

/*
func (r *Resources) Update(ctx context.Context) error {
	resources, err := r.app.factory.GetDiscoveryClient().APIResources(ctx, "main")
	if err != nil {
		return err
	}
	r.view.SetTitle(fmt.Sprintf("  %s [%d]  ", r.title, len(resources.Spec.Groups)))
	// data
	sort.Slice(resources.Spec.Groups, func(i, j int) bool {
		return getResourceID(resources.Spec.Groups[i]) < getResourceID(resources.Spec.Groups[j])
	})

	resourceIDs := []string{}
	for _, resource := range resources.Spec.Groups {
		resourceID := getResourceID(resource)
		resourceIDs = append(resourceIDs, resourceID)
		row := r.FindResourcesRow(resourceID)
		r.SetResourcesRow(row, resource)
	}

	CleanTableRows(r.DataTable, func(row int) bool {
		return slices.Contains(resourceIDs, r.GetResourcesRow(row))
	})

	// header
	columnWidth := CalculateMaxWidths(r.DataTable)
	SetTableHeader(r.HeaderTable, columnWidth, "GroupResource", "Version", "Kind", "ListKind", "Namespaced", "Categories")

	r.HeaderTable.ScrollToBeginning()
	r.DataTable.ScrollToBeginning()
	r.app.ForceDraw()
	return nil
}
*/

func getResourceID(resource *discoverypb.APIResource) string {
	return schema.GroupResource{Group: resource.Group, Resource: resource.Resource}.String()
}

func GetResourceRowData(resource *discoverypb.APIResource) []string {
	row := []string{}
	row = append(row, getResourceID(resource))
	row = append(row, resource.Version)
	row = append(row, resource.Kind)
	row = append(row, resource.ListKind)
	row = append(row, fmt.Sprintf("%t", resource.Namespaced))
	row = append(row, fmt.Sprintf("%v", resource.Categories))
	return row
}

/*
func (r *Resources) SetResourcesRow(row int, resource *discoverypb.APIResource) {
	style := tcell.StyleDefault.
		Foreground(tcell.ColorLightSkyBlue).
		Background(tcell.ColorBlack)

	r.DataTable.SetCell(row, 0, tview.NewTableCell(getResourceID(resource)).SetReference(resource).SetStyle(style))
	r.DataTable.SetCell(row, 1, tview.NewTableCell(resource.Version).SetStyle(style))
	r.DataTable.SetCell(row, 2, tview.NewTableCell(resource.Kind).SetStyle(style))
	r.DataTable.SetCell(row, 3, tview.NewTableCell(resource.ListKind).SetStyle(style))
	r.DataTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%t", resource.Namespaced)).SetStyle(style))
	r.DataTable.SetCell(row, 5, tview.NewTableCell(fmt.Sprintf("%v", resource.Categories)).SetStyle(style))
}

func (r *Resources) GetResourcesRow(row int) string {
	resource := r.DataTable.GetCell(row, 0).GetReference().(*discoverypb.APIResource)
	return getResourceID(resource)
}

func (r *Resources) FindResourcesRow(resourceID string) int {
	rowCount := r.DataTable.GetRowCount()
	for row := 0; row < rowCount; row++ {
		if resourceID == r.GetResourcesRow(row) {
			return row
		}
	}
	return rowCount
}
*/

func (r *Resources) actvateChildPage(ctx context.Context, row, column int) error {
	resource, ok := r.DataTable.GetCell(row, column).GetReference().(*discoverypb.APIResource)
	if !ok {
		return fmt.Errorf("%s addChildPage resource not available in row %d, column %d", r.name, row, column)
	}
	p := NewResource(ctx, r, resource)
	p.ActivatePage(ctx)
	return nil
}

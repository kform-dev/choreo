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
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

type Resource struct {
	HeaderTable *tview.Table
	DataTable   *tview.Table

	parent   Page
	app      *App
	name     string
	title    string
	apiGroup *discoverypb.APIResource
	view     *tview.Flex
	style    tcell.Style
	cancel   func()
}

func NewResource(ctx context.Context, parent Page, apiGroup *discoverypb.APIResource) Page {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}

	r := &Resource{
		name:        getResourceID(apiGroup),
		title:       fmt.Sprintf("Resource: %s  ", getResourceID(apiGroup)),
		HeaderTable: tview.NewTable(),
		DataTable:   tview.NewTable(),
		app:         app,
		apiGroup:    apiGroup,
		style: tcell.StyleDefault.
			Foreground(tcell.ColorLightSkyBlue).
			Background(tcell.ColorBlack),
		parent: parent,
	}

	r.SetTable(ctx)
	// TODO Mousecapture
	r.SetView()

	app.pages.AddPage(r.name, r.view, true, true)
	return r
}

func (r *Resource) SetTable(ctx context.Context) {
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

func (r *Resource) SetView() {
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

func (r *Resource) RegisterPageAction(ctx context.Context) {
	// this would register local keys of the page to help actions

}

func (r *Resource) DeActivatePage(ctx context.Context) {
	r.parent.ActivatePage(ctx)
	r.StopStream(ctx)
}

func (r *Resource) ActivatePage(ctx context.Context) {
	// init action
	r.app.actions = NewKeyActions()
	r.app.header.InitPageAction()
	r.app.pages.RegisterPageAction(ctx) // central main page keys

	// activate page action
	r.app.header.ActivatePageAction("")

	// activate local page
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
func (r *Resource) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	return event // Propagate to global handler
}

func (r *Resource) StopStream(ctx context.Context) {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Resource) StartStream(ctx context.Context) {
	if r.cancel != nil {
		r.cancel()
	}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   r.apiGroup.Group,
		Version: r.apiGroup.Version,
		Kind:    r.apiGroup.Kind,
	})
	rspch := r.app.factory.GetResourceClient().Watch(ctx, u, &resourceclient.ListOptions{
		Proxy:             r.app.factory.GetProxy(),
		Branch:            r.app.factory.GetBranch(),
		ShowManagedFields: false,
		ExprSelector:      &resourcepb.ExpressionSelector{},
	})
	r.DataTable.Clear()
	idset := NewIdentifierSet()

	SetTableHeader(r.HeaderTable, []int{10}, "Name")

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
				case resourcepb.Watch_ADDED, resourcepb.Watch_MODIFIED:
					obj := map[string]any{}
					if err := yaml.Unmarshal(event.Object, &obj); err != nil {
						return
					}
					u := &unstructured.Unstructured{}
					u.Object = obj

					id, exists := idset.AddIdentifier(GetResourceInstanceID(u))
					if exists {
						UpdateRowAt(r.DataTable, id, u, GetResourceInstanceRowData(u), r.style)
					} else {
						InsertRowAt(r.DataTable, id, u, GetResourceInstanceRowData(u), r.style)
					}

				case resourcepb.Watch_DELETED:
					id := idset.DeleteIdentifier(GetResourceInstanceID(u))
					if id != -1 {
						DeleteRow(r.DataTable, id)
					}
				default:

				}
				columnWidth := CalculateMaxWidths(r.DataTable)
				SetTableHeader(r.HeaderTable, columnWidth, "Name")
				r.HeaderTable.ScrollToBeginning()
				r.DataTable.ScrollToBeginning()
				//r.app.ForceDraw()
			})

		}
	}
}

func GetResourceInstanceID(u *unstructured.Unstructured) string {
	return u.GetName()
}

func GetResourceInstanceRowData(u *unstructured.Unstructured) []string {
	row := []string{}
	row = append(row, u.GetName())
	return row
}

/*
func (r *Resource) Update(ctx context.Context) error {
	ul := &unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(r.apiGroup.GVK())
	err := r.app.factory.GetResourceClient().List(ctx, ul, resourceclient.ListOptions{
		ExprSelector: &resourcepb.ExpressionSelector{},
		Branch:       "main",
	})
	if err != nil {
		return err
	}

	r.view.SetTitle(fmt.Sprintf("  %s [%d]  ", r.title, len(ul.Items)))
	// header
	//SetTableHeader(r.HeaderTable, "Name")
	// data
	sort.Slice(ul.Items, func(i, j int) bool {
		return getResourceInstanceID(&ul.Items[i]) < getResourceInstanceID(&ul.Items[j])
	})

	resourceInstanceIDs := []string{}
	for _, u := range ul.Items {
		resourceInstanceID := getResourceInstanceID(&u)
		resourceInstanceIDs = append(resourceInstanceIDs, resourceInstanceID)
		row := r.FindResourceInstanceRow(resourceInstanceID)
		r.SetResourceInstanceRow(row, &u)
	}

	CleanTableRows(r.DataTable, func(row int) bool {
		return slices.Contains(resourceInstanceIDs, r.GetResourceInstanceRow(row))
	})

	// header
	columnWidth := CalculateMaxWidths(r.DataTable)
	SetTableHeader(r.HeaderTable, columnWidth, "Name")

	r.HeaderTable.ScrollToBeginning()
	r.DataTable.ScrollToBeginning()
	r.app.ForceDraw()
	return nil
}
*/

/*
func (r *Resource) SetResourceInstanceRow(row int, u *unstructured.Unstructured) {
	style := tcell.StyleDefault.
		Foreground(tcell.ColorLightSkyBlue).
		Background(tcell.ColorBlack)
	r.DataTable.SetCell(row, 0, tview.NewTableCell(getResourceInstanceID(u)).SetReference(u).SetStyle(style))
}

func (r *Resource) GetResourceInstanceRow(row int) string {
	u := r.DataTable.GetCell(row, 0).GetReference().(*unstructured.Unstructured)
	return getResourceInstanceID(u)
}

func (r *Resource) FindResourceInstanceRow(resourceID string) int {
	rowCount := r.DataTable.GetRowCount()
	for row := 0; row < rowCount; row++ {
		if resourceID == r.GetResourceInstanceRow(row) {
			return row
		}
	}
	return rowCount
}
*/

func (r *Resource) actvateChildPage(ctx context.Context, row, column int) error {
	u, ok := r.DataTable.GetCell(row, column).GetReference().(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("%s addChildPage resource not available in row %d, column %d", r.name, row, column)
	}
	p := NewResourceInstance(ctx, r, u)
	p.ActivatePage(ctx)
	//p.InitAction(ctx)
	//p.ActivateAction(ctx, "")
	return nil
}

package view

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Resources struct {
	HeaderTable *tview.Table
	DataTable   *tview.Table

	app   *App
	name  string
	title string
	view  *tview.Flex
}

func NewResources(ctx context.Context, parent Page) {
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
	}
	app.views[r.DataTableName()] = r.DataTable
	app.views[r.HeaderTableName()] = r.HeaderTable

	style := tcell.StyleDefault.
		Foreground(tcell.ColorLightSkyBlue).
		Background(tcell.ColorBlack)

	r.HeaderTable.SetBackgroundColor(tcell.ColorBlack)
	r.HeaderTable.SetFixed(1, 0)
	r.DataTable.SetBackgroundColor(tcell.ColorBlack)
	r.DataTable.SetFixed(0, 0)
	r.DataTable.
		SetSelectable(true, false).
		Select(0, 0).
		SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(app.views[parent.Name()])
			if r.view != nil {
				r.view.SetTitle(r.title)
			}
			CleanTable(r.DataTable)
			r.app.ticker.stop()
		}).
		SetSelectedFunc(func(row, column int) {
			// function that is creating the new page
			r.actvateChildPage(ctx, row, column)

		})
	// TODO Mousecapture
	view := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.HeaderTable, 1, 1, false).
		AddItem(r.DataTable, 0, 1, true)
	view.
		SetBorder(true).
		SetTitle(r.title).
		SetBlurFunc(func() {
			if app.ticker != nil {
				app.ticker.stop()
				app.ticker = nil
			}
		}).
		SetBorderStyle(style)
	r.view = view
	app.Main.AddPage(r.name, view, true, true)
	app.menu.AddItem(r.title, "", 'r', func() {
		r.Activate(ctx)
	}).SetMainTextColor(tcell.ColorWhite)
}

func (r *Resources) Name() string {
	return r.name
}

func (r *Resources) HeaderTableName() string {
	return fmt.Sprintf("%s-header", r.name)
}

func (r *Resources) DataTableName() string {
	return fmt.Sprintf("%s-data", r.name)
}

func (r *Resources) Activate(ctx context.Context) {
	r.app.Main.SwitchToPage(r.name)
	r.app.SetFocus(r.DataTable)
	r.app.ForceDraw()
	r.app.ticker = NewTicker(r.app.Application, 2*time.Second, func() {
		r.Update(ctx)
	})
	r.app.ticker.start()
}

func (r *Resources) Update(ctx context.Context) error {
	resources, err := r.app.factory.GetDiscoveryClient().APIResources(ctx, "main")
	if err != nil {
		return err
	}
	r.view.SetTitle(fmt.Sprintf("  %s [%d]  ", r.title, len(resources)))
	// header
	//columnWidth := SetTableHeader(r.HeaderTable, "GroupResource", "Version", "Kind", "ListKind", "Namespaced", "Categories")
	// data
	sort.Slice(resources, func(i, j int) bool {
		return getResourceID(resources[i]) < getResourceID(resources[j])
	})

	resourceIDs := []string{}
	for _, resource := range resources {
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

func getResourceID(resource *discoverypb.APIResource) string {
	return schema.GroupResource{Group: resource.Group, Resource: resource.Resource}.String()
}

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

func (r *Resources) actvateChildPage(ctx context.Context, row, column int) error {
	resource, ok := r.DataTable.GetCell(row, column).GetReference().(*discoverypb.APIResource)
	if !ok {
		return fmt.Errorf("%s addChildPage resource not available in row %d, column %d", r.name, row, column)
	}
	NewResource(ctx, r, resource)
	return nil
}

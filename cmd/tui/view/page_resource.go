package view

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Resource struct {
	HeaderTable *tview.Table
	DataTable   *tview.Table

	app      *App
	name     string
	title    string
	apiGroup *discoverypb.APIResource
	view     *tview.Flex
}

func NewResource(ctx context.Context, parent Page, apiGroup *discoverypb.APIResource) {
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
	}
	app.views[r.DataTableName()] = r.DataTable
	app.views[r.HeaderTableName()] = r.HeaderTable

	style := tcell.StyleDefault.
		Foreground(tcell.ColorLightSkyBlue).
		Background(tcell.ColorBlack)

	r.HeaderTable.SetBackgroundColor(tcell.ColorBlack)
	r.DataTable.SetBackgroundColor(tcell.ColorBlack)
	r.DataTable.
		SetSelectable(true, false).
		Select(0, 0).
		SetDoneFunc(func(key tcell.Key) {
			parent.Activate(ctx)
			delete(app.views, r.HeaderTableName())
			delete(app.views, r.DataTableName())
		}).
		SetSelectedFunc(func(row, column int) {
			// function that is creating the new page
			r.actvateChildPage(ctx, row, column)
		}).
		ScrollToBeginning()
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
	r.Activate(ctx)
}

func (r *Resource) Name() string {
	return r.name
}

func (r *Resource) HeaderTableName() string {
	return fmt.Sprintf("%s-header", r.name)
}

func (r *Resource) DataTableName() string {
	return fmt.Sprintf("%s-data", r.name)
}

func (r *Resource) Activate(ctx context.Context) {
	r.app.Main.SwitchToPage(r.name)
	r.app.SetFocus(r.DataTable)
	r.app.ForceDraw()
	r.app.ticker = NewTicker(r.app.Application, 2*time.Second, func() {
		r.Update(ctx)
	})
	r.app.ticker.start()
}

func (r *Resource) Update(ctx context.Context) error {
	ul := &unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(schema.GroupVersionKind{Group: r.apiGroup.Group, Version: r.apiGroup.Version, Kind: r.apiGroup.Kind})
	err := r.app.factory.GetResourceClient().List(ctx, ul, &resourceclient.ListOptions{
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

func getResourceInstanceID(u *unstructured.Unstructured) string {
	return u.GetName()
}

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

func (r *Resource) actvateChildPage(ctx context.Context, row, column int) error {
	u, ok := r.DataTable.GetCell(row, column).GetReference().(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("%s addChildPage resource not available in row %d, column %d", r.name, row, column)
	}
	NewResourceInstance(ctx, r, u)
	return nil
}

package view

import (
	"context"

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
func NewInfo(ctx context.Context) *Info {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}

	r := &Info{
		Table: tview.NewTable(),
		name:  "info",
		title: "Info",
		app:   app,
	}
	app.views[r.name] = r.Table
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

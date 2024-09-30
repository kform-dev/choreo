package view

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Menu struct {
	*tview.List

	app   *App
	name  string
	title string
}

// NewInfo creates a new Info view.
func NewMenu(ctx context.Context) *Menu {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}
	r := &Menu{
		List:  tview.NewList(),
		name:  "menu",
		title: "menu",
		app:   app,
	}
	app.menu = r
	app.views[r.name] = r.List

	//app.menu.AddItem("main", "", '0', r.setSelectedbranch("main"))
	//app.menu.AddItem("1", "", '1', r.setSelectedbranch("main"))
	app.menu.AddItem("Quit", "", 'q', app.Stop)
	//r.SetBorder(true)
	//r.SetTitle("menu")
	r.ShowSecondaryText(false)
	r.SetBackgroundColor(tcell.ColorBlack)
	r.SetDoneFunc(r.app.Stop)

	return r
}

func (r *Menu) Name() string {
	return r.name
}

func (r *Menu) Activate(ctx context.Context) {
	r.app.Main.SwitchToPage(r.name)
	r.app.SetFocus(r.List)
	r.app.ForceDraw()

}

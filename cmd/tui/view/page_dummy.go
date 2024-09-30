package view

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Dummy struct {
	*tview.TextView

	app   *App
	name  string
	title string
	text  string
}

func NewDummy(ctx context.Context) error {
	app, err := extractApp(ctx)
	if err != nil {
		return err
	}

	r := &Dummy{
		name:     "dummy",
		title:    "Dummy",
		TextView: tview.NewTextView(),
		app:      app,
	}
	app.views[r.name] = r.TextView

	r.TextView.
		SetDynamicColors(true).
		SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(app.menu)
		}).
		//SetText("dummy").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			//fmt.Printf("Key in dummy pressed: %v\n", event.Key())
			if event.Key() == tcell.KeyESC {
				r.app.SetFocus(app.menu)
				return nil
			}
			return event
		})
	r.TextView.SetBackgroundColor(tcell.ColorBlack)

	// Mouseapture todo
	view := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.TextView, 0, 1, true)
	view.
		SetBorder(true).
		SetTitle(r.title).
		SetBlurFunc(func() {
			if app.ticker != nil {
				app.ticker.stop()
				app.ticker = nil
			}
		})
	app.Main.AddPage(r.name, view, true, true)
	app.menu.AddItem(r.title, "", 'd', func() {
		r.Activate(ctx)
		/*
			app.SwitchToPage(r.name, "") // focus is set in the switch to page fn
			app.ticker = NewTicker(app.Application, 2*time.Second, func() {
				r.Update(ctx)
			})
			app.ticker.start()
		*/
	}).SetMainTextColor(tcell.ColorWhite)

	return nil
}

func (r *Dummy) Activate(ctx context.Context) {
	r.app.Main.SwitchToPage(r.name)
	r.app.SetFocus(r.TextView)
	r.app.ForceDraw()
	r.app.ticker = NewTicker(r.app.Application, 2*time.Second, func() {
		r.Update(ctx)
	})
	r.app.ticker.start()
}

func (r *Dummy) Update(ctx context.Context) error {
	switch r.text {
	case "a":
		r.text = "b"
	case "b":
		r.text = "a"
	default:
		r.text = "a"
	}
	r.TextView.SetText(r.text)
	//r.app.ForceDraw()
	return nil
}

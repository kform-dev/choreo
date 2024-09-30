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
	view  *tview.Flex
	style tcell.Style
}

func NewDummy(ctx context.Context, pages *Pages) Page {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}

	r := &Dummy{
		name:     "dummy",
		title:    "Dummy",
		TextView: tview.NewTextView(),
		app:      app,
		style: tcell.StyleDefault.
			Foreground(tcell.ColorLightSkyBlue).
			Background(tcell.ColorBlack),
	}
	r.SetTextView(ctx)
	// Mouseapture todo
	r.SetView()
	pages.AddPage(r.name, r.view, true, true)
	return r
}

func (r *Dummy) SetTextView(ctx context.Context) {
	r.TextView.
		SetDynamicColors(true).
		SetDoneFunc(func(key tcell.Key) {
			r.DeActivatePage(ctx)
		}).
		SetTextStyle(r.style)
}

func (r *Dummy) SetView() {
	view := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.TextView, 0, 1, true)
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

func (r *Dummy) RegisterPageAction(ctx context.Context) {
	ka := KeyAction{
		Key:       KeyD,
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

func (r *Dummy) TviewPrimitive() tview.Primitive {
	return r.TextView
}
func (r *Dummy) DeActivatePage(ctx context.Context) {
}

func (r *Dummy) ActivatePage(ctx context.Context) {
	r.app.actions = NewKeyActions()
	r.app.header.InitPageAction()
	r.app.pages.RegisterPageAction(ctx)

	r.app.header.ActivatePageAction("main")

	r.app.pages.SwitchToPage(r.name)
	r.app.SetFocus(r.TextView)
	r.app.ticker = NewTicker(r.app.Application, 2*time.Second, func() {
		r.Update(ctx)
	})
	r.app.ticker.start()
	r.app.ForceDraw()
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

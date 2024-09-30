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

var LogoSmall = []string{
	`  ____ _   _  ___  ____  _____ ___   `,
	` / ___| | | |/ _ \|  _ \| ____/ _ \  `,
	`| |   | |_| | | | | |_) |  _|| | | | `,
	`| |___|  _  | |_| |  _ <| |__| |_| | `,
	` \____|_| |_|\___/|_| \_\_____\___/  `,
}

type Logo struct {
	*tview.TextView

	app   *App
	name  string
	title string
}

// NewLogo returns a new logo.
func NewLogo(app *App) *Logo {
	r := Logo{
		TextView: tview.NewTextView(),
		name:     LogoString(),
		title:    "Logo",
		app:      app,
	}
	r.SetText(r.name)
	r.SetDynamicColors(true)
	r.SetTextColor((tcell.ColorOrange))
	r.SetBackgroundColor(tcell.ColorBlack)
	//r.SetBorder(true)
	//r.SetTitle(r.title)

	return &r
}

func LogoString() string {
	logoText := ""
	for _, line := range LogoSmall {
		logoText += line + "\n"
	}
	return logoText
}

func logoWidth() int {
	maxWidth := 0
	for _, line := range LogoSmall {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

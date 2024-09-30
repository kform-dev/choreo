package view

import (
	"context"

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
func NewLogo(ctx context.Context) *Logo {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}
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
	//r.SetTitle("Logo")

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

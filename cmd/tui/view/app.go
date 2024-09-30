package view

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/rivo/tview"
)

type App struct {
	factory util.Factory

	*tview.Application

	Main   *tview.Pages // this might have to become a stack
	views  map[string]tview.Primitive
	menu   *Menu
	ticker *ticker
	//branchSelection  string
	checkedOutBranch string
}

func NewApp(f util.Factory) *App {
	app := &App{
		factory:     f,
		Application: tview.NewApplication(),
		Main:        tview.NewPages(),
		views:       make(map[string]tview.Primitive),
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			}
		}
		return event
	})

	return app
}

func (a *App) Init(ctx context.Context) error {
	ctx = context.WithValue(ctx, KeyApp, a)
	// initialize the header
	menu := NewMenu(ctx)
	info := NewInfo(ctx)
	logo := NewLogo(ctx)
	flexHeader := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(info, 0, 1, true).
		AddItem(menu, 0, 2, true).
		AddItem(logo, logoWidth(), 1, false)

	// initialize Main resource pages
	NewDummy(ctx)
	NewResources(ctx, menu) // menu acts a s a parent

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(flexHeader, 6, 1, true). // 4 sets min height to 4
		AddItem(a.Main, 0, 6, true)

	a.SetRoot(layout, true)
	a.SetFocus(a.views["menu"])

	//os.Exit(1)

	return nil
}

// Run starts the application loop.
func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	//go a.watchBranch(ctx)

	if err := a.Application.Run(); err != nil {
		return err
	}
	return nil
}

func (a *App) watchBranch(ctx context.Context) {
	rspch := a.factory.GetBranchClient().Watch(ctx, &branchpb.Watch_Request{Options: &branchpb.Watch_Options{}})

	for {
		select {
		case <-ctx.Done():
			// watch stoppped
			return
		case event, ok := <-rspch:
			if !ok {
				return
			}
			switch event.EventType {
			case branchpb.Watch_ADDED, branchpb.Watch_MODIFIED:
				if event.BranchObj.CheckedOut {
					a.checkedOutBranch = event.BranchObj.Name
				}
			case branchpb.Watch_DELETED:
			default:

			}
		}
	}
}

func (a *App) stop(evt *tcell.EventKey) *tcell.EventKey {
	if a.ticker != nil {
		a.ticker.stop()
	}
	a.Stop()
	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// AsKey converts rune to keyboard key.,.
func AsKey(evt *tcell.EventKey) tcell.Key {
	if evt.Key() != tcell.KeyRune {
		return evt.Key()
	}
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}

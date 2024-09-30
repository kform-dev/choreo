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
	"github.com/rivo/tview"
)

type Header struct {
	*tview.Flex

	branchMenu *BranchMenu
	cmdMenu    *CommandMenu
	info       *Info
	logo       *Logo
}

func NewHeader(app *App) *Header {
	// initialize the header
	branchMenu := NewBranchMenu(app)
	cmdMenu := NewCommandMenu(app)
	info := NewInfo(app)
	logo := NewLogo(app)

	return &Header{
		Flex:       tview.NewFlex(),
		info:       info,
		branchMenu: branchMenu,
		cmdMenu:    cmdMenu,
		logo:       logo,
	}
}

func (r *Header) SwitchLayout(context string) {
	// clear flex
	r.Clear()
	switch context {
	case "main":
		r.SetDirection(tview.FlexColumn).
			AddItem(r.info, 0, 1, false).
			AddItem(r.branchMenu, 0, 1, false).
			AddItem(r.cmdMenu, 0, 1, false).
			AddItem(r.logo, logoWidth(), 1, false) // logoWidth example fixed to 10
	default:
		r.SetDirection(tview.FlexColumn).
			AddItem(r.info, 0, 1, false).
			AddItem(r.cmdMenu, 0, 2, false).
			AddItem(r.logo, logoWidth(), 1, false) // logoWidth example fixed to 10
	}
}

func (r *Header) InitPageAction() {
	r.branchMenu.InitPageAction()
	r.cmdMenu.InitPageAction()
}

func (r *Header) ActivatePageAction(context string) {
	r.SwitchLayout(context)
	r.branchMenu.ActivatePageAction()
	r.cmdMenu.ActivatePageAction()
}

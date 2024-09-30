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
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type ResourceInstance struct {
	*tview.TextView

	parent    Page
	app       *App
	name      string
	title     string
	u         *unstructured.Unstructured
	view      *tview.Flex
	yamlStyle YAMLStyle
	style     tcell.Style
}

func NewResourceInstance(ctx context.Context, parent Page, u *unstructured.Unstructured) Page {
	app, err := extractApp(ctx)
	if err != nil {
		panic(err)
	}

	r := &ResourceInstance{
		name:     u.GetName(),
		title:    fmt.Sprintf("  ResourceInstance: %s.%s  ", schema.FromAPIVersionAndKind(u.GetAPIVersion(), u.GetKind()).String(), u.GetName()),
		TextView: tview.NewTextView(),
		app:      app,
		u:        u,
		yamlStyle: YAMLStyle{
			//KeyColor:   tcell.ColorLightSkyBlue,
			KeyColor: tcell.NewRGBColor(86, 156, 214),
			//ValueColor: tcell.ColorTeal,
			ValueColor: tcell.NewRGBColor(206, 145, 120),
			ColonColor: tcell.ColorWhite,
		},
		style: tcell.StyleDefault.
			Foreground(tcell.ColorLightSkyBlue).
			Background(tcell.ColorBlack),
		parent: parent,
	}
	r.SetTextView(ctx)
	// TODO Mousecapture
	r.SetView()
	app.pages.AddPage(r.name, r.view, true, true)
	return r
}

func (r *ResourceInstance) SetTextView(ctx context.Context) {
	r.TextView.
		SetDynamicColors(true).
		SetText(colorizeYAML(r.yamlStyle, YamlToString(r.u))).
		SetDoneFunc(func(key tcell.Key) {
			r.DeActivatePage(ctx)
		}).
		SetTextStyle(r.style)
}

func (r *ResourceInstance) SetView() {
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

func (r *ResourceInstance) RegisterPageAction(ctx context.Context) {
	// this would register local keys of the page to help actions
}

func (r *ResourceInstance) DeActivatePage(ctx context.Context) {
	r.parent.ActivatePage(ctx)
	r.app.ticker.stop()
}

func (r *ResourceInstance) ActivatePage(ctx context.Context) {
	// init action
	r.app.actions = NewKeyActions()
	r.app.header.InitPageAction()
	r.app.pages.RegisterPageAction(ctx)

	// activate page action
	r.app.header.ActivatePageAction("") //

	// activate local page
	r.app.pages.SwitchToPage(r.name)
	r.app.SetFocus(r.TextView)
	r.TextView.SetInputCapture(r.HandleInput)
	r.app.ticker = NewTicker(r.app.Application, 2*time.Second, func() {
		r.Update(ctx)
	})
	r.app.ticker.start()
	r.app.ForceDraw()
}

// In your page implementation
func (r *ResourceInstance) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	return event // Propagate to global handler
}

func (r *ResourceInstance) Update(ctx context.Context) error {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.FromAPIVersionAndKind(r.u.GetAPIVersion(), r.u.GetKind()))
	err := r.app.factory.GetResourceClient().Get(ctx, types.NamespacedName{Namespace: r.u.GetNamespace(), Name: r.u.GetName()}, u, &resourceclient.GetOptions{
		Branch:            "main",
		ShowManagedFields: true,
	})
	if err != nil {
		// todo log error
		return err
	}

	r.TextView.SetText(colorizeYAML(r.yamlStyle, YamlToString(u)))
	//r.app.ForceDraw()
	return nil
}

var (
	keyValRX = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\/|\s]+):\s(.+)\z`)
	keyRX    = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\/|\s]+):\s*\z`)
)

const (
	yamlFullFmt  = "%s[key::b]%s[colon::-]: [val::]%s"
	yamlKeyFmt   = "%s[key::b]%s[colon::-]:"
	yamlValueFmt = "[val::]%s"
)

func enableRegion(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, "<<<", "["), ">>>", "]")
}

type YAMLStyle struct {
	KeyColor   tcell.Color `json:"keyColor" yaml:"keyColor"`
	ValueColor tcell.Color `json:"valueColor" yaml:"valueColor"`
	ColonColor tcell.Color `json:"colonColor" yaml:"colonColor"`
}

func colorizeYAML(style YAMLStyle, raw string) string {
	lines := strings.Split(tview.Escape(raw), "\n")
	fullFmt := strings.Replace(yamlFullFmt, "[key", "["+style.KeyColor.String(), 1)
	fullFmt = strings.Replace(fullFmt, "[colon", "["+style.ColonColor.String(), 1)
	fullFmt = strings.Replace(fullFmt, "[val", "["+style.ValueColor.String(), 1)

	keyFmt := strings.Replace(yamlKeyFmt, "[key", "["+style.KeyColor.String(), 1)
	keyFmt = strings.Replace(keyFmt, "[colon", "["+style.ColonColor.String(), 1)

	valFmt := strings.Replace(yamlValueFmt, "[val", "["+style.ValueColor.String(), 1)

	buff := make([]string, 0, len(lines))
	for _, l := range lines {
		res := keyValRX.FindStringSubmatch(l)
		if len(res) == 4 {
			buff = append(buff, enableRegion(fmt.Sprintf(fullFmt, res[1], res[2], res[3])))
			continue
		}

		res = keyRX.FindStringSubmatch(l)
		if len(res) == 3 {
			buff = append(buff, enableRegion(fmt.Sprintf(keyFmt, res[1], res[2])))
			continue
		}

		buff = append(buff, enableRegion(fmt.Sprintf(valFmt, l)))
	}

	return strings.Join(buff, "\n")
}

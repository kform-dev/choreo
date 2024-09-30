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

	app       *App
	name      string
	title     string
	u         *unstructured.Unstructured
	yamlStyle YAMLStyle
}

func NewResourceInstance(ctx context.Context, parent Page, u *unstructured.Unstructured) {
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
	}
	app.views[r.name] = r.TextView

	style := tcell.StyleDefault.
		Foreground(tcell.ColorLightSkyBlue).
		Background(tcell.ColorBlack)

	r.TextView.
		SetDynamicColors(true).
		SetText(colorizeYAML(r.yamlStyle, YamlToString(u))).
		SetDoneFunc(func(key tcell.Key) {
			parent.Activate(ctx)
			delete(app.views, r.name)
			r.app.ticker.stop()
		}).
		SetTextStyle(style)

	// TODO Mousecapture
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
		}).
		SetBorderStyle(style)
	app.Main.AddPage(r.name, view, true, true)
	r.Activate(ctx)
}

func (r *ResourceInstance) Name() string {
	return r.name
}

func (r *ResourceInstance) Activate(ctx context.Context) {
	r.app.Main.SwitchToPage(r.name)
	r.app.SetFocus(r.TextView)
	r.app.ForceDraw()
	r.app.ticker = NewTicker(r.app.Application, 2*time.Second, func() {
		r.Update(ctx)
	})
	r.app.ticker.start()
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
	r.app.ForceDraw()
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

/*
func colorizeYAML(node interface{}, indent int) string {
	// Use hexadecimal or color names
	keyColor := "[orange]"
	valueColor := "[skyblue]"
	output := ""
	prefix := strings.Repeat("  ", indent)
	switch node := node.(type) {
	case map[string]interface{}:
		for k, v := range node {
			output += fmt.Sprintf("%s%s%s:[white]: ", prefix, keyColor, k)
			output += colorizeYAML(v, indent+1) + "\n"
		}
	case []interface{}:
		for _, v := range node {
			output += colorizeYAML(v, indent+1) + "\n"
		}
	default:
		output = fmt.Sprintf("%s%v[white]", valueColor, node)
	}
	return output
}
*/

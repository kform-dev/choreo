package view

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func CleanTable(table *tview.Table) {
	row := 0
	for row < table.GetRowCount() {
		table.RemoveRow(row)
	}
}

func CleanTableRows(table *tview.Table, exists func(row int) bool) {
	row := 0
	for row < table.GetRowCount() {
		if exists(row) {
			row++
		} else {
			table.RemoveRow(row)
		}
	}
}

func SetTableHeader(table *tview.Table, maxColumnWidth []int, headers ...string) {
	//if table.GetRowCount() == 0 {
	for column, header := range headers {
		width := len(header)
		if len(maxColumnWidth) != 0 {
			width = maxColumnWidth[column]
		}
		table.SetCell(0, column, NewHeaderTableCell(fmt.Sprintf("%-*s", width, header)))
	}
	//}
}

func NewHeaderTableCell(text string) *tview.TableCell {
	style := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack).
		Attributes(tcell.AttrNone)

	return tview.NewTableCell(text).
		SetSelectable(false).
		SetStyle(style)
}

func YamlToString(u *unstructured.Unstructured) string {
	b, err := yaml.Marshal(u.Object)
	if err != nil {
		// todo log error
		return ""
	}
	return string(b)
}

func CalculateMaxWidths(table *tview.Table) []int {
	maxCols := table.GetColumnCount()
	maxWidths := make([]int, maxCols)

	for row := 0; row < table.GetRowCount(); row++ {
		for col := 0; col < maxCols; col++ {
			cell := table.GetCell(row, col)
			length := len(cell.Text)
			if length > maxWidths[col] {
				maxWidths[col] = length
			}
		}
	}
	return maxWidths
}

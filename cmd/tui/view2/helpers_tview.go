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
	"fmt"
	"strings"

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
		table.SetCell(0, column, NewHeaderTableCell(fmt.Sprintf("%-*s", width, strings.ToUpper(header))))
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

func SetMenuRow(table *tview.Table, row int, key, val string, keyStyle, valStyle tcell.Style) {
	table.SetCell(row, 0, tview.NewTableCell(key).SetStyle(keyStyle))
	table.SetCell(row, 1, tview.NewTableCell(val).SetStyle(valStyle))
}

// InsertRowAt inserts a new row into the table at the specified position
func UpdateRowAt(table *tview.Table, index int, ref any, rowData []string, style tcell.Style) {
	// Insert new row
	for col, data := range rowData {
		if col == 0 {
			table.SetCell(index, col, tview.NewTableCell(data).SetReference(ref).SetStyle(style))
		} else {
			table.SetCell(index, col, tview.NewTableCell(data).SetStyle(style))
		}
	}
}

// InsertRowAt inserts a new row into the table at the specified position
func InsertRowAt(table *tview.Table, index int, ref any, rowData []string, style tcell.Style) {
	// Move existing rows down one position from the index
	rowCount := table.GetRowCount()
	for row := rowCount - 1; row >= index; row-- {
		for col := 0; col < table.GetColumnCount(); col++ {
			cell := table.GetCell(row, col)
			table.SetCell(row+1, col, cell)
		}
	}

	// Insert new row
	for col, data := range rowData {
		if col == 0 {
			table.SetCell(index, col, tview.NewTableCell(data).SetReference(ref).SetStyle(style))
		} else {
			table.SetCell(index, col, tview.NewTableCell(data).SetStyle(style))
		}
	}
}

// DeleteRow removes a row from the table at the specified position
func DeleteRow(table *tview.Table, index int) {
	rowCount := table.GetRowCount()
	// Move all rows up from the index
	for row := index; row < rowCount-1; row++ {
		for col := 0; col < table.GetColumnCount(); col++ {
			cell := table.GetCell(row+1, col)
			table.SetCell(row, col, cell)
		}
	}
	// Remove the last row which is now duplicated
	table.RemoveRow(rowCount - 1)
}

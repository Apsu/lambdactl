package ui

import (
	lg "github.com/charmbracelet/lipgloss"
	lgt "github.com/charmbracelet/lipgloss/table"
)

func NewTable(headers TableRow) *Table {
	t := &Table{
		headers: headers,
		cursor:  1,
		offset:  0,
		table:   lgt.New(),
	}

	t.table.Headers(headers...)
	t.Rows(TableRows{})

	return t
}

func (t *Table) SetStyles(tableStyle, borderStyle, selectStyle lg.Style) {
	t.tableStyle = tableStyle
	t.borderStyle = borderStyle
	t.selectStyle = selectStyle
	t.table.BorderStyle(borderStyle)
	t.table.StyleFunc(t.styleFunc)
}

func (t *Table) View() string {
	return t.table.Render()
}

func (t *Table) styleFunc(row, col int) lg.Style {
	style := t.tableStyle

	// Highlight the selected row
	if row == t.cursor-1 {
		style = t.selectStyle
	}

	return style
}

func (t *Table) Resize(width, height int) {
	t.width = width
	t.height = height

	// TODO: Calculate from table border style
	t.viewport = t.height

	// Clamp in case terminal is reeeeally short
	if t.viewport < 0 {
		t.viewport = 0
	}

	// Update underlying table size
	t.table.Width(t.width)
	t.table.Height(t.viewport)

	// // Adjust offset if necessary
	// if t.offset+t.viewport > t.length {
	// 	t.offset = t.length - t.viewport
	// 	if t.offset < 0 {
	// 		t.offset = 0
	// 	}
	// }

	// Refill rows
	t.table.Offset(t.offset)
	t.Rows(t.rows)

	// // Adjust cursor if it's out of view
	// if t.cursor < max(t.offset, 1) {
	// 	t.cursor = t.offset
	// } else if t.cursor > t.offset+t.viewport {
	// 	t.cursor = t.offset + t.viewport
	// }
}

func (t *Table) MoveUp(step int) {
	if t.cursor > 1 {
		t.cursor--
		if t.cursor == t.offset {
			t.offset--
			t.table.Offset(t.offset)
			t.Rows(t.rows)
		}
	}
}

func (t *Table) MoveTop() {

}

func (t *Table) MoveDown(step int) {
	if t.cursor < t.length {
		t.cursor++
		if t.cursor > t.offset+t.viewport {
			t.offset++
			t.cursor = t.offset + t.viewport
			t.table.Offset(t.offset)
			t.Rows(t.rows)
		}
	}
}

func (t *Table) MoveBottom() {

}

func (t *Table) ScrollUp() {

}

func (t *Table) ScrollTop() {

}

func (t *Table) ScrollDown() {

}

func (t *Table) ScrollBottom() {

}

func (t *Table) Rows(filledRows TableRows) {
	// Store filled rows for refills
	t.rows = filledRows
	t.length = len(filledRows)

	// This allows table to fill viewport when not full
	var displayRows TableRows
	// Pad rows to viewport if less filled than offset viewport
	if t.length < t.offset+t.viewport {
		displayRows = append(filledRows, make(TableRows, t.offset+t.viewport-t.length)...)
		// Or trim to offset viewport
	} else {
		displayRows = filledRows[:t.offset+t.viewport]
	}

	// Convert for underlying lipgloss table
	converted := make([][]string, len(displayRows))
	for i, row := range displayRows {
		converted[i] = []string(row)
	}

	// Replace lipgloss table data
	t.table.Data(lgt.NewStringData(converted...))
}

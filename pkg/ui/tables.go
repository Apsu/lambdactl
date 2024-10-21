package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

func (m *Model) updateColumns() {
	columns := m.content.Columns()
	rows := m.content.Rows()

	if m.width == 0 || len(columns) == 0 {
		return
	}

	minWidth := 4 // Define a minimum width for each column
	numCols := len(columns)

	// Calculate total available width excluding padding
	totalWidth := m.width - numCols*2

	// Initialize max widths (headers and rows)
	maxWidths := make([]int, numCols)
	sumWidths := 0

	// Calculate maxWidths based on headers and row contents
	for i, col := range columns {
		maxWidths[i] = len(col.Title)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > maxWidths[i] {
				maxWidths[i] = len(cell)
			}
		}
	}

	// Calculate the sum of maxWidths
	for _, width := range maxWidths {
		sumWidths += width
	}

	// Check if the totalWidth is less than the sum of the min widths
	minSumWidths := numCols * minWidth
	if totalWidth < minSumWidths {
		// If totalWidth is smaller than the sum of min widths, return early with min widths
		for i := range maxWidths {
			maxWidths[i] = minWidth
		}
	} else {
		// Calculate the delta between current width and target totalWidth
		delta := totalWidth - sumWidths

		// Evenly expand or contract widths based on delta
		for delta != 0 {
			for i := range maxWidths {
				if delta > 0 {
					// Expand evenly until totalWidth is reached
					maxWidths[i]++
					delta--
					sumWidths++
				} else if delta < 0 {
					// Contract evenly until totalWidth is reached
					maxWidths[i]--
					delta++
					sumWidths--
				}
				if delta == 0 {
					break
				}
			}
		}
	}

	// Set adjusted widths back to columns
	for i, width := range maxWidths {
		columns[i].Width = width
	}

	m.content.SetColumns(columns)
}

func (m *Model) loadInstanceTable() {
	// Clear table
	m.content.SetRows([]table.Row{})

	// Setup headers
	columns := []table.Column{
		{Title: "Name"},
		{Title: "Region"},
		{Title: "Public IP"},
		{Title: "Private IP"},
		{Title: "Model"},
		{Title: "GPUs"},
		{Title: "CPUs"},
		{Title: "RAM"},
		{Title: "Storage"},
		{Title: "Status"},
	}

	// Setup data
	rows := make([]table.Row, len(m.instances))
	for i, instance := range m.instances {
		model := instance.InstanceType.GPUDescription
		if model == "N/A" {
			model = instance.InstanceType.Description
		}
		rows[i] = table.Row{
			instance.Name,
			instance.Region.Name,
			instance.IP,
			instance.PrivateIP,
			model,
			fmt.Sprintf("%d", instance.InstanceType.Specs.GPUs),
			fmt.Sprintf("%d", instance.InstanceType.Specs.VCPUs),
			fmt.Sprintf("%d GiB", instance.InstanceType.Specs.MemoryGiB),
			fmt.Sprintf("%d GiB", instance.InstanceType.Specs.StorageGiB),
			instance.Status,
		}
	}

	// Add to table
	m.content.SetColumns(columns)
	m.content.SetRows(rows)

	// Resize columns
	m.updateColumns()
}

func (m *Model) loadOptionTable() {
	// Clear table
	m.content.SetRows([]table.Row{})

	// Setup headers
	columns := []table.Column{
		{Title: "Region"},
		{Title: "Model"},
		{Title: "GPUs"},
		{Title: "vCPUs"},
		{Title: "Memory"},
		{Title: "Storage"},
		{Title: "$/Hour"},
	}

	// Setup data
	rows := make([]table.Row, len(m.options))
	for i, option := range m.options {
		model := option.Type.GPUDescription
		if model == "N/A" {
			model = option.Type.Description
		}
		rows[i] = table.Row{
			option.Region,
			model,
			fmt.Sprintf("%d", option.Type.Specs.GPUs),
			fmt.Sprintf("%d", option.Type.Specs.VCPUs),
			fmt.Sprintf("%d GiB", option.Type.Specs.MemoryGiB),
			fmt.Sprintf("%d GiB", option.Type.Specs.StorageGiB),
			fmt.Sprintf("%.2f", float64(option.Type.PriceCentsPerHour)/100),
		}
	}

	// Add to table
	m.content.SetColumns(columns)
	m.content.SetRows(rows)

	// Resize columns
	m.updateColumns()
}

func (m *Model) loadFilesystemTable() {
	// Clear table
	m.content.SetRows([]table.Row{})

	// Setup headers
	columns := []table.Column{
		{Title: "Name"},
		{Title: "Region"},
		{Title: "Mount Point"},
		{Title: "Bytes Used"},
		{Title: "In Use"},
		{Title: "Created By"},
		{Title: "Created At"},
	}

	// Setup data
	rows := make([]table.Row, len(m.filesystems))
	for i, fs := range m.filesystems {
		rows[i] = []string{
			fs.Name,
			fs.Region.Name,
			fs.MountPoint,
			fmt.Sprintf("%d", fs.BytesUsed),
			fmt.Sprintf("%t", fs.IsInUse),
			fs.CreatedBy.Email,
			fs.Created.Format(time.RFC3339),
		}
	}

	// Add to table
	m.content.SetColumns(columns)
	m.content.SetRows(rows)

	// Resize columns
	m.updateColumns()
}

func (m *Model) loadSSHTable() {
	// Clear table
	m.content.SetRows([]table.Row{})

	// Setup headers
	columns := []table.Column{
		{Title: "Name"},
		{Title: "Public Key"},
	}

	// Setup data
	rows := make([]table.Row, len(m.sshKeys))
	for i, sshKey := range m.sshKeys {
		rows[i] = []string{
			sshKey.Name,
			fmt.Sprintf("%.150s", sshKey.PublicKey),
		}
	}

	// Add to table
	m.content.SetColumns(columns)
	m.content.SetRows(rows)

	// Resize columns
	m.updateColumns()
}

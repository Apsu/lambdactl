package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

func (m *Model) loadInstanceTable() {
	log.Println("Loading instance table, instances: ", m.instances)
	// Add headers to the table
	m.content.SetColumns(
		[]table.Column{
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
		},
	)

	// Add instance data
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
			fmt.Sprintf("%d", instance.InstanceType.Specs.MemoryGiB),
			fmt.Sprintf("%d", instance.InstanceType.Specs.StorageGiB),
			instance.Status,
		}
	}

	m.content.SetRows(rows)
}

func (m *Model) loadOptionTable() {
	log.Println("Loading option table, options: ", m.options)
	// Add headers to the table
	m.content.SetColumns(
		[]table.Column{
			{Title: "Region"},
			{Title: "Model"},
			{Title: "GPUs"},
			{Title: "vCPUs"},
			{Title: "Memory"},
			{Title: "Storage"},
			{Title: "$/Hour"},
		},
	)

	// Add instance data
	for _, instance := range m.instances {
		model := instance.InstanceType.GPUDescription
		if model == "N/A" {
			model = instance.InstanceType.Description
		}
		m.content.SetRows([]table.Row{
			[]string{
				instance.Region.Name,
				model,
				fmt.Sprintf("%d", instance.InstanceType.Specs.GPUs),
				fmt.Sprintf("%d", instance.InstanceType.Specs.VCPUs),
				fmt.Sprintf("%d", instance.InstanceType.Specs.MemoryGiB),
				fmt.Sprintf("%d", instance.InstanceType.Specs.StorageGiB),
				fmt.Sprintf("%.2f", float64(instance.InstanceType.PriceCentsPerHour)/100),
			},
		})
	}
}

func (m *Model) loadFilesystemTable() {
	log.Println("Loading filesystem table, filesystems: ", m.filesystems)
	// Add headers to the table
	m.content.SetColumns(
		[]table.Column{
			{Title: "Name"},
			{Title: "Region"},
			{Title: "Mount Point"},
			{Title: "Bytes Used"},
			{Title: "In Use"},
			{Title: "Created By"},
			{Title: "Created At"},
		},
	)

	// Add filesystem data
	for _, fs := range m.filesystems {
		m.content.SetRows([]table.Row{
			[]string{
				fs.Name,
				fs.Region.Name,
				fs.MountPoint,
				fmt.Sprintf("%d", fs.BytesUsed),
				fmt.Sprintf("%t", fs.IsInUse),
				fs.CreatedBy.Email,
				fs.Created.Format(time.RFC3339),
			},
		})
	}
}

func (m *Model) loadSSHTable() {
	log.Println("Loading SSH key table, SSH keys: ", m.sshKeys)
	// Add headers to the table
	m.content.SetColumns(
		[]table.Column{
			{Title: "Name"},
			{Title: "Public Key"},
		},
	)

	// Add SSH key data
	for _, ssh := range m.sshKeys {
		m.content.SetRows([]table.Row{
			[]string{
				ssh.Name,
				ssh.PublicKey,
			},
		})
	}
}

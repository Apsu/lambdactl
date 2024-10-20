package ui

import (
	"fmt"
	"sort"

	"lambdactl/pkg/api"
)

func instancesToRows(instances []api.Instance) TableRows {
	// Sort the slice before transforming into table rows
	sort.SliceStable(instances, func(i, j int) bool {
		// TODO: Figure out custom sorting by column index
		// Takes custom sort function sortBy
		// return sortBy(machines[i], machines[j])

		// Sort by name for now
		return instances[i].Name < instances[j].Name
	})

	rows := make(TableRows, len(instances))
	for i, instance := range instances {
		model := instance.InstanceType.GPUDescription
		if model == "N/A" {
			model = instance.InstanceType.Description
		}
		rows[i] = TableRow{
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

	return rows
}

func optionsToRows(options []api.InstanceOption) TableRows {
	// Sort the slice before transforming into table rows
	sort.SliceStable(options, func(i, j int) bool {
		// TODO: Figure out custom sorting by column index
		// Takes custom sort function sortBy
		// return sortBy(machines[i], machines[j])

		// Sort by price for now
		return options[i].Type.PriceCentsPerHour < options[j].Type.PriceCentsPerHour
	})

	rows := make(TableRows, len(options))
	for i, option := range options {
		model := option.Type.GPUDescription
		if model == "N/A" {
			model = option.Type.Description
		}
		rows[i] = TableRow{
			option.Region,
			model,
			fmt.Sprintf("%d", option.Type.Specs.GPUs),
			fmt.Sprintf("%d", option.Type.Specs.VCPUs),
			fmt.Sprintf("%d GiB", option.Type.Specs.MemoryGiB),
			fmt.Sprintf("%d GiB", option.Type.Specs.StorageGiB),
			fmt.Sprintf("%.2f", float64(option.Type.PriceCentsPerHour)/100),
		}
	}
	return rows
}

package ui

import (
	"fmt"
	"strings"
	"time"

	"lambdactl/pkg/api"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func NewModel() *Model {
	client := api.NewAPIClient(
		viper.GetString("apiUrl"),
		viper.GetString("apiKey"),
	)

	runningColumns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Region", Width: 15},
		{Title: "Public IP", Width: 15},
		{Title: "Private IP", Width: 15},
		{Title: "Model", Width: 5},
		{Title: "Bus", Width: 10},
		{Title: "GPUs", Width: 5},
		{Title: "vCPUs", Width: 5},
		{Title: "Memory", Width: 10},
		{Title: "Status", Width: 10},
	}

	optionColumns := []table.Column{
		{Title: "Region", Width: 15},
		{Title: "GPUs", Width: 5},
		{Title: "Model", Width: 5},
		{Title: "Bus", Width: 10},
		{Title: "$/Hour", Width: 20},
	}

	runningTable := table.New(
		table.WithColumns(runningColumns),
		table.WithFocused(true),
	)

	optionTable := table.New(
		table.WithColumns(optionColumns),
		table.WithFocused(false),
	)

	return &Model{
		client:          client,
		refreshInterval: 30 * time.Second,
		currentState:    runningState,
		previousState:   runningState,
		activeTable:     &runningTable,
		runningTable:    runningTable,
		optionTable:     optionTable,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshInstances(), m.refreshOptions(), m.startTimer())
}

func (m Model) startTimer() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return timerMsg{}
	})
}

func (m Model) refreshInstances() tea.Cmd {
	return func() tea.Msg {
		instances, err := m.client.ListInstances()
		if err != nil {
			return errMsg{err}
		}

		for i, instance := range instances {
			if parsed, err := api.ParseInstanceType(instance.InstanceType); err == nil {
				instance.InstanceType.Specs = parsed
				instances[i] = instance
			}
		}

		return instancesMsg{instances}
	}
}

func (m Model) refreshOptions() tea.Cmd {
	return func() tea.Msg {
		options, err := m.client.FetchInstanceOptions()
		if err != nil {
			return errMsg{err}
		}

		for i, option := range options {
			if parsed, err := api.ParseOptionType(option.Type); err == nil {
				option.Specs = parsed
				options[i] = option
			}
		}

		return optionsMsg{options}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	switch msg := msg.(type) {
	case errMsg:
		m.errorMsg = msg.Error()
		return m, nil
	// case createdVMMsg:
	// 	return m, m.refreshInstances()
	// case filterMsg:
	// 	m.filter = msg.filter
	// 	return m, m.applyFilter()
	case instancesMsg:
		m.machines = msg.instances
		m.runningTable.SetRows(machineSliceToTableRows(m.machines))
		return m, nil
	case optionsMsg:
		m.options = msg.options
		m.optionTable.SetRows(optionSliceToTableRows(m.options))
		return m, nil
	case timerMsg:
		return m, tea.Batch(m.refreshInstances(), m.refreshOptions(), m.startTimer())
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", "right", "l":
			if m.currentState == runningState {
				m.selectedMachine = &m.machines[m.activeTable.Cursor()]
				m.currentState = detailState
			}
		case "esc", "left", "h":
			if m.currentState != runningState {
				m.currentState = runningState
				m.activeTable = &m.runningTable
			}
		case "s":
			if m.currentState == detailState {
				program.ReleaseTerminal()
				if err := m.SSHCmd(); err != nil {
					fmt.Printf("SSH result: %v", err)
				}
				program.RestoreTerminal()
			}
		case "n":
			if m.currentState == runningState {
				m.currentState = optionState
				m.activeTable = &m.optionTable
			}
			// return m, m.optionCmd()
			// case "ctrl+l":
			// 	return m, m.refreshInstances()
		}
	}

	// Let the table handle key events
	t, cmd := m.activeTable.Update(msg)
	*m.activeTable = t
	return m, cmd
}

func (m Model) View() string {
	if m.errorMsg != "" {
		return "Error: " + m.errorMsg
	}

	// Dispatch to state's view func
	switch m.currentState {
	case runningState, optionState:
		return tableView(m)
	default:
		return fmt.Sprintf("Unknown view for state: %v", m.currentState)
	}
}

func tableView(m Model) string {
	var b strings.Builder

	b.WriteString("VM List\n\n")
	b.WriteString(m.activeTable.View())
	b.WriteString("\nFilter: " + m.filter + "\n" +
		"(q) Quit, (n) New VM, (enter) Details, (/) Filter")

	return b.String()
}

func detailView(m Model) string {
	var b strings.Builder

	machineYAML, err := yaml.Marshal(m.selectedMachine)
	if err != nil {
		return fmt.Sprintf("error marshalling YAML: %v", err)
	}

	b.WriteString("VM Details\n\n")
	b.WriteString(string(machineYAML))
	b.WriteString("\n(esc) Back, (s) SSH")

	return b.String()
}

func optionView(m Model) string {
	var b strings.Builder

	b.WriteString("Instances Available\n\n")
	b.WriteString(m.optionTable.View())
	b.WriteString("\nFilter: " + m.filter + "\n" +
		"(esc) Back, (enter) option, (/) Filter")

	return b.String()
}

func (m Model) optionCmd() tea.Cmd {
	return func() tea.Msg {
		requested := api.InstanceOption{
			Region: "us-south-2",
			Specs: api.InstanceSpecs{
				Bus:   "sxm5",
				GPUs:  1,
				Model: "h100",
			},
		}
		options, err := m.client.FetchInstanceOptions()
		if err != nil {
			return errMsg{err}
		}
		bestOption, err := api.SelectBestInstanceOption(options, requested)
		if err != nil {
			return errMsg{err}
		}
		_, err = m.client.LaunchInstances(bestOption, 1)
		if err != nil {
			return errMsg{err}
		}
		return m.refreshInstances()
	}
}

func (m Model) SSHCmd() error {
	err := m.client.SSHIntoMachine(*m.selectedMachine)
	if err != nil {
		return err
	}

	return nil
}

// func (m Model) applyFilter() tea.Cmd {
// 	return func() tea.Msg {
// 		filtered, err := m.client.ListInstances()
// 		if err != nil {
// 			return errMsg{err}
// 		}
// 		m.machines = filtered
// 		m.table.SetRows(machineSliceToTableRows(m.machines))
// 		return nil
// 	}
// }

func machineSliceToTableRows(machines []api.InstanceDetails) []table.Row {
	rows := make([]table.Row, len(machines))
	for i, machine := range machines {
		instanceSpecs, err := api.ParseInstanceType(machine.InstanceType)
		if err != nil {
			instanceSpecs.Model = machine.InstanceType.Name
		}
		rows[i] = table.Row{
			machine.Name,
			machine.Region.Name,
			machine.IP,
			machine.PrivateIP,
			instanceSpecs.Model,
			instanceSpecs.Bus,
			fmt.Sprintf("%d", machine.InstanceType.Specs.GPUs),
			fmt.Sprintf("%d", machine.InstanceType.Specs.VCPUs),
			fmt.Sprintf("%d GiB", machine.InstanceType.Specs.MemoryGiB),
			machine.Status,
		}
	}
	return rows
}

func optionSliceToTableRows(options []api.InstanceOption) []table.Row {
	rows := make([]table.Row, len(options))
	for i, option := range options {
		rows[i] = table.Row{
			option.Region,
			fmt.Sprintf("%d", option.Specs.GPUs),
			option.Specs.Model,
			option.Specs.Bus,
			fmt.Sprintf("%.2f", float64(option.PriceHour)/100),
		}
	}
	return rows
}

// Global so we can manipulate terminal/stdin reader for SSH
var program *tea.Program

func Start() error {
	program = tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := program.Run()
	if err != nil {
		return fmt.Errorf("error running program: %v", err)
	}
	return nil
}

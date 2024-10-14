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

	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Region", Width: 10},
		{Title: "IP", Width: 15},
		{Title: "Private IP", Width: 15},
		{Title: "Type", Width: 15},
		{Title: "vCPUs", Width: 5},
		{Title: "Memory", Width: 10},
		{Title: "GPUs", Width: 5},
		{Title: "Status", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	return &Model{
		client:          client,
		refreshInterval: 10 * time.Second,
		currentState:    listState,
		previousState:   listState,
		table:           t,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshInstances(), m.startTimer())
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
		return instancesMsg{instances}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case errMsg:
		m.errorMsg = msg.Error()
		return m, nil
	case createdVMMsg:
		return m, m.refreshInstances()
	case filterMsg:
		m.filter = msg.filter
		return m, m.applyFilter()
	case instancesMsg:
		m.machines = msg.instances
		m.table.SetRows(machineSliceToTableRows(m.machines))
		return m, nil
	case timerMsg:
		return m, tea.Batch(m.refreshInstances(), m.startTimer())
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", "right", "l":
			if m.currentState.Name == "list" {
				m.selectedMachine = &m.machines[m.table.Cursor()]
				m.currentState = detailState
			}
		case "esc", "left", "h":
			if m.currentState.Name == "detail" {
				m.currentState = listState
			}
		case "s":
			if m.selectedMachine != nil {
				program.ReleaseTerminal()
				if err := m.SSHCmd(); err != nil {
					fmt.Printf("SSH result: %v", err)
				}
				program.RestoreTerminal()
			}
		case "n":
			return m, m.launchCmd()
			// case "ctrl+l":
			// 	return m, m.refreshInstances()
		}
	}

	// Let the table handle key events
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.errorMsg != "" {
		return "Error: " + m.errorMsg
	}

	// Dispatch to state's view func
	return m.currentState.View(m)
}

func listView(m Model) string {
	var b strings.Builder

	b.WriteString("VM List\n\n")
	b.WriteString(m.table.View())
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

func (m Model) SSHCmd() error {
	err := m.client.SSHIntoMachine(*m.selectedMachine)
	if err != nil {
		return err
	}

	return nil
}

func launchView(m Model) string {
	return ""
}

func (m Model) launchCmd() tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.LaunchInstances(api.GPUOption{}, 1)
		if err != nil {
			return errMsg{err}
		}
		return m.refreshInstances()
	}
}

func (m Model) applyFilter() tea.Cmd {
	return func() tea.Msg {
		filtered, err := m.client.ListInstances()
		if err != nil {
			return errMsg{err}
		}
		m.machines = filtered
		m.table.SetRows(machineSliceToTableRows(m.machines))
		return nil
	}
}

func machineSliceToTableRows(machines []api.InstanceDetails) []table.Row {
	rows := make([]table.Row, len(machines))
	for i, machine := range machines {
		rows[i] = table.Row{
			machine.Name,
			machine.Region.Name,
			machine.IP,
			machine.PrivateIP,
			machine.InstanceType.Name,
			fmt.Sprintf("%d", machine.InstanceType.Specs.VCPUs),
			fmt.Sprintf("%d GiB", machine.InstanceType.Specs.MemoryGiB),
			fmt.Sprintf("%d", machine.InstanceType.Specs.GPUs),
			machine.Status,
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

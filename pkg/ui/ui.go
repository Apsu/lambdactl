package ui

import (
	"fmt"
	"strings"
	"time"

	"lambdactl/pkg/api"
	// "lambdactl/pkg/utils"
	"gopkg.in/yaml.v3"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type Model struct {
	client          *api.APIClient
	machines        []api.InstanceDetails
	selectedMachine *api.InstanceDetails
	filter          string
	table           table.Model
	currentView     string
	previousView    string
	errorMsg        string
	refreshInterval time.Duration
}

func NewModel() Model {
	client := api.NewAPIClient(viper.GetString("apiUrl"), viper.GetString("apiKey"))

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

	return Model{
		client:          client,
		refreshInterval: 10 * time.Second,
		currentView:     "list",
		table:           t,
	}
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type filterMsg struct {
	filter string
}

type createdVMMsg struct{}

type instancesMsg struct {
	instances []api.InstanceDetails
}

type sshCompleteMsg struct {
	err error
}

type timerMsg struct{}

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
	// case createdVMMsg:
	// 	return m, m.refreshInstances()
	// case filterMsg:
	// 	m.filter = msg.filter
	// 	return m, m.applyFilter()
	case instancesMsg:
		m.machines = msg.instances
		m.table.SetRows(machineSliceToTableRows(m.machines))
		return m, nil
	// case sshCompleteMsg:
	// 	m.currentView = m.previousView
	// 	return m, nil
	case timerMsg:
		return m, tea.Batch(m.refreshInstances(), m.startTimer())
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", "right", "l":
			if m.currentView == "list" {
				m.selectedMachine = &m.machines[m.table.Cursor()]
				m.currentView = "details"
			}
		case "esc", "left", "h":
			if m.currentView == "details" {
				m.currentView = "list"
			}
			// case "s":
			// 	if m.selectedMachine != nil {
			// 		m.previousView = m.currentView
			// 		m.currentView = "ssh"
			// 		return m, m.StartSSHSession()
			// 	}
			// 	return m, nil
			// case "n":
			// 	return m, m.createVM()
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

	// Apply the appropriate view rendering
	switch m.currentView {
	case "list":
		return m.listView()
	case "details":
		return m.detailView()
	case "ssh":
		return "SSH session in progress"
	default:
		return "Unknown View"
	}
}

func (m Model) listView() string {
	var b strings.Builder

	b.WriteString("VM List\n\n")
	b.WriteString(m.table.View())
	b.WriteString("\nFilter: " + m.filter + "\n" +
		"(q) Quit, (n) New VM, (enter) Details, (/) Filter")

	return b.String()
}

func (m Model) detailView() string {
	var b strings.Builder
	b.WriteString("VM Details\n\n")
	// Marshal machine to YAML to pretty print
	machineYAML, _ := yaml.Marshal(m.selectedMachine)
	b.WriteString(string(machineYAML))
	b.WriteString("\n(b) Back, (s) SSH")

	return b.String()
}

func (m Model) createVM() tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.LaunchInstances(api.GPUOption{}, 1)
		if err != nil {
			return errMsg{err}
		}
		return m.refreshInstances()
	}
}

func (m Model) StartSSHSession() tea.Cmd {
	return func() tea.Msg {
		err := m.client.SSHIntoMachine(*m.selectedMachine)
		if err != nil {
			return sshCompleteMsg{err}
		}

		return sshCompleteMsg{}
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

func Start() (tea.Model, error) {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	return p.Run()
}

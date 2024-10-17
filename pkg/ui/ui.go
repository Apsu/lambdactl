package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"lambdactl/pkg/api"
	"lambdactl/pkg/sshlib"
	"lambdactl/pkg/utils"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

func NewModel() *Model {
	client := api.NewAPIClient(
		viper.GetString("api-url"),
		viper.GetString("api-key"),
	)

	styles := table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(draculaHeaderColor)).
			Background(lipgloss.Color(draculaBackground)).
			Padding(0, 1),
		Cell: lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(draculaForeground)).
			Background(lipgloss.Color(draculaHighlight)),
	}

	runningColumns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Region", Width: 20},
		{Title: "Public IP", Width: 15},
		{Title: "Private IP", Width: 15},
		{Title: "Model", Width: 25},
		{Title: "GPUs", Width: 5},
		{Title: "vCPUs", Width: 5},
		{Title: "Memory", Width: 8},
		{Title: "Storage", Width: 10},
		{Title: "Status", Width: 10},
	}

	optionColumns := []table.Column{
		{Title: "Region", Width: 20},
		{Title: "Model", Width: 25},
		{Title: "GPUs", Width: 5},
		{Title: "vCPUs", Width: 5},
		{Title: "Memory", Width: 8},
		{Title: "Storage", Width: 10},
		{Title: "$/Hour", Width: 10},
	}

	runningTable := table.New(
		table.WithColumns(runningColumns),
		table.WithFocused(true),
	)
	runningTable.SetStyles(styles)

	optionTable := table.New(
		table.WithColumns(optionColumns),
		table.WithFocused(false),
	)
	optionTable.SetStyles(styles)

	return &Model{
		client:          client,
		refreshInterval: 30 * time.Second,
		errorTimeout:    5 * time.Second,
		currentState:    runningState,
		runningTable:    runningTable,
		optionTable:     optionTable,
		// launchForm:      *launchForm,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshInstances(),
		m.refreshOptions(),
		m.startTimer(m.refreshInterval, timerMsg{}),
	)
}

func (m Model) startTimer(interval time.Duration, msg any) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return msg
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

func (m Model) refreshOptions() tea.Cmd {
	return func() tea.Msg {
		options, err := m.client.FetchInstanceOptions()
		if err != nil {
			return errMsg{err}
		}

		return optionsMsg{options}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, m.startTimer(m.errorTimeout, clearErrMsg{})
		}
		return m, nil
	case clearErrMsg:
		m.errorMsg = ""
	}
	switch m.currentState {
	case runningState:
		return m.updateRunningState(msg)
	case optionState:
		return m.updateOptionState(msg)
	case detailState:
		return m.updateDetailState(msg)
	case launchState:
		return m.updateLaunchState(msg)
	default:
		return m, nil
	}
}

func (m Model) updateRunningState(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case instancesMsg:
		m.machines = msg.instances
		m.runningTable.SetRows(machineSliceToTableRows(m.machines))
		return m, nil
	case timerMsg:
		return m, tea.Batch(m.refreshInstances(), m.startTimer(m.refreshInterval, timerMsg{}))
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.selectedMachine = &m.machines[m.runningTable.Cursor()]
			m.currentState = detailState
			return m, m.startTimer(m.errorTimeout, timerMsg{})
		case "tab":
			m.currentState = optionState
			m.optionTable.Focus()
			return m, m.refreshOptions()
		case "e":
			return m, tea.Cmd(
				func() tea.Msg {
					return errMsg{fmt.Errorf("THIS IS A TEST")}
				},
			)
		}
	}

	var cmd tea.Cmd
	m.runningTable, cmd = m.runningTable.Update(msg)
	return m, cmd
}

func (m Model) updateOptionState(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case optionsMsg:
		m.options = msg.options
		m.optionTable.SetRows(optionSliceToTableRows(m.options))
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.currentState = runningState
			m.runningTable.Focus()
			return m, tea.Batch(
				m.refreshInstances(),
				m.startTimer(m.refreshInterval, timerMsg{}),
			)
		case "enter":
			m.currentState = launchState
			m.selectedOption = &m.options[m.optionTable.Cursor()]
		case "r", "ctrl+l":
			return m, m.refreshOptions()
		}
	}

	var cmd tea.Cmd
	m.optionTable, cmd = m.optionTable.Update(msg)
	return m, cmd
}

func (m Model) updateDetailState(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timerMsg:
		return m, m.startTimer(m.errorTimeout, timerMsg{})
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentState = runningState
			m.runningTable.Focus()
			return m, tea.Batch(
				m.refreshInstances(),
				m.startTimer(m.refreshInterval, timerMsg{}),
			)
		case "s":
			return m, tea.Exec(
				&sshlib.SSHExecCommand{
					Target: sshlib.SSHTarget{
						Host:    m.selectedMachine.IP,
						KeyName: "id_aap",
						Port:    22,
						User:    "ubuntu",
					},
				}, func(err error) tea.Msg { return errMsg{err} },
			)
		}
	}

	return m, nil
}

func (m Model) updateLaunchState(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentState = optionState
			m.optionTable.Focus()
		case "l":
			m.selectedOption = &m.options[m.optionTable.Cursor()]
			m.currentState = runningState
			m.runningTable.Focus()
			return m, tea.Sequence(
				m.launchCmd(),
				m.refreshInstances(),
				m.startTimer(m.refreshInterval, timerMsg{}),
			)
		}
	}

	return m, nil
}

func (m Model) View() string {
	var viewContent, errorDisplay string
	if m.errorMsg != "" {
		errorDisplay = errorStyle.Render("Error: " + m.errorMsg)
	}

	// Dispatch to state's view func
	switch m.currentState {
	case runningState:
		viewContent = runningView(m)
	case optionState:
		viewContent = optionView(m)
	case detailState:
		viewContent = detailView(m)
	case launchState:
		viewContent = launchView(m)
	default:
		viewContent = fmt.Sprintf("Unknown view for state: %v", m.currentState)
	}

	return lipgloss.JoinVertical(lipgloss.Top, viewContent, errorDisplay)
}

func runningView(m Model) string {
	var b strings.Builder

	b.WriteString("VM List\n\n")
	b.WriteString(borderStyle.Render(m.runningTable.View()))
	b.WriteString("\n\n(q) Quit (enter) View Details (tab) Launch Options")
	return b.String()
}

func optionView(m Model) string {
	var b strings.Builder

	b.WriteString("Instances Available\n\n")
	b.WriteString(borderStyle.Render(m.optionTable.View()))
	b.WriteString("\n\n(q) Quit (enter) View Option (tab) Running Instances (r) Refresh Options")

	return b.String()
}

func detailView(m Model) string {
	var b strings.Builder

	b.WriteString("VM Details\n\n")
	b.WriteString(borderStyle.Render(utils.PrettyYAML(m.selectedMachine)))
	b.WriteString("\n\n(q) Quit (esc) Back (s) SSH")

	return b.String()
}

func launchView(m Model) string {
	var b strings.Builder

	b.WriteString("Launch this option?\n\n")
	b.WriteString(utils.PrettyYAML(m.selectedOption))
	b.WriteString("\n\n(q) Quit (esc) Back (l) Launch")

	return b.String()
}

func (m Model) launchCmd() tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.LaunchInstances(*m.selectedOption, 1)
		if err != nil {
			return errMsg{err}
		}
		return m.refreshInstances()
	}
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
	// Sort the slice before transforming into table rows
	sort.SliceStable(machines, func(i, j int) bool {
		// Takes custom sort function sortBy
		// return sortBy(machines[i], machines[j])

		// Sort by name for now
		return machines[i].Name < machines[j].Name
	})

	rows := make([]table.Row, len(machines))
	for i, machine := range machines {
		model := machine.InstanceType.GPUDescription
		if model == "N/A" {
			model = machine.InstanceType.Description
		}
		rows[i] = table.Row{
			machine.Name,
			machine.Region.Name,
			machine.IP,
			machine.PrivateIP,
			model,
			fmt.Sprintf("%d", machine.InstanceType.Specs.GPUs),
			fmt.Sprintf("%d", machine.InstanceType.Specs.VCPUs),
			fmt.Sprintf("%d GiB", machine.InstanceType.Specs.MemoryGiB),
			fmt.Sprintf("%d GiB", machine.InstanceType.Specs.StorageGiB),
			machine.Status,
		}
	}

	return rows
}

func optionSliceToTableRows(options []api.InstanceOption) []table.Row {
	// Sort the slice before transforming into table rows
	sort.SliceStable(options, func(i, j int) bool {
		// Takes custom sort function sortBy
		// return sortBy(machines[i], machines[j])

		// Sort by price for now
		return options[i].Type.PriceCentsPerHour < options[j].Type.PriceCentsPerHour
	})

	rows := make([]table.Row, len(options))
	for i, option := range options {
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
	return rows
}

func Start() error {
	program := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := program.Run()
	if err != nil {
		return fmt.Errorf("error running program: %v", err)
	}
	return nil
}

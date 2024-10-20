package ui

import (
	"fmt"
	"lambdactl/pkg/sshlib"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
	case tea.WindowSizeMsg: // Called once on startup, then on SIGWINCH signals
		m.width = msg.Width
		m.height = msg.Height

		// TODO: Make this better
		tableHeight := m.height - lipgloss.Height(headerStyle.Render("")) - lipgloss.Height(footerStyle.Render("")) // Adjust height

		m.instanceTable.Resize(m.width, tableHeight)
		m.optionTable.Resize(m.width, tableHeight)

		return m, nil
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
		m.instances = msg.instances
		m.instanceTable.Rows(instancesToRows(m.instances))
		return m, nil
	case timerMsg:
		return m, tea.Batch(m.refreshInstances(), m.startTimer(m.refreshInterval, timerMsg{}))
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.instanceTable.MoveUp(1)
		case "down", "j":
			m.instanceTable.MoveDown(1)
		case "enter":
			m.selectedInstance = m.instances[m.instanceTable.cursor-1]
			m.currentState = detailState
			// TODO: Clear errors?
			return m, nil //m.startTimer(m.errorTimeout, timerMsg{})
		case "tab":
			m.currentState = optionState
			// TODO: Add refresh timer?
			return m, m.refreshOptions()
		case "e":
			return m, tea.Cmd(
				func() tea.Msg {
					return errMsg{fmt.Errorf("THIS IS A TEST")}
				},
			)
		}
	}

	return m, nil
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
		case "s":
			return m, tea.Exec(
				&sshlib.SSHExecCommand{
					Target: sshlib.SSHTarget{
						Host:    m.selectedInstance.IP,
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

func (m Model) updateOptionState(msg tea.Msg) (tea.Model, tea.Cmd) {
	// switch msg := msg.(type) {
	// case optionsMsg:
	// 	m.options = msg.options
	// 	m.optionTable.SetRows(optionSliceToTableRows(m.options))
	// 	return m, nil
	// case tea.KeyMsg:
	// 	switch keypress := msg.String(); keypress {
	// 	case "q", "ctrl+c":
	// 		return m, tea.Quit
	// 	case "tab":
	// 		m.currentState = runningState
	// 		m.runningTable.Focus()
	// 		return m, tea.Batch(
	// 			m.refreshInstances(),
	// 			m.startTimer(m.refreshInterval, timerMsg{}),
	// 		)
	// 	case "enter":
	// 		m.currentState = launchState
	// 		m.selectedOption = &m.options[m.optionTable.Cursor()]
	// 	case "r", "ctrl+l":
	// 		return m, m.refreshOptions()
	// 	}
	// }

	// var cmd tea.Cmd
	// m.optionTable, cmd = m.optionTable.Update(msg)
	// return m, cmd
	return m, nil
}

func (m Model) updateLaunchState(msg tea.Msg) (tea.Model, tea.Cmd) {
	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	switch keypress := msg.String(); keypress {
	// 	case "q", "ctrl+c":
	// 		return m, tea.Quit
	// 	case "esc":
	// 		m.currentState = optionState
	// 		m.optionTable.Focus()
	// 	case "l":
	// 		m.selectedOption = &m.options[m.optionTable.Cursor()]
	// 		m.currentState = runningState
	// 		m.runningTable.Focus()
	// 		return m, tea.Sequence(
	// 			m.launchCmd(),
	// 			m.refreshInstances(),
	// 			m.startTimer(m.refreshInterval, timerMsg{}),
	// 		)
	// 	}
	// }

	return m, nil
}

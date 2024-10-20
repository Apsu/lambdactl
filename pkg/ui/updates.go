package ui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("Update: ", msg)
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

		// Adjust table height
		// tableHeight := m.height - headerStyle.GetHeight() - footerStyle.GetHeight()

		// m.content.SetHeight(tableHeight)
		// m.content.SetWidth(m.width)

		return m, nil
	}

	switch m.currentState {
	case instanceState:
		return m.updateInstanceState(msg)
	case optionState:
		return m.updateOptionState(msg)
	case filesystemState:
		return m.updateFilesystemState(msg)
	case sshState:
		return m.updateSSHState(msg)
	default:
		return m, nil
	}
}

func (m *Model) updateInstanceState(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("UpdateInstanceState: ", msg)
	switch msg := msg.(type) {
	case instancesMsg:
		m.instances = msg.instances
		m.loadInstanceTable()
		return m, nil
	case timerMsg:
		return m, tea.Batch(m.refreshInstances(), m.startTimer(m.refreshInterval, timerMsg{}))
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.cursor = m.content.Cursor()
			m.detailsPanel = true
			return m, nil
		case "tab":
			m.currentState = optionState
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
	m.content, cmd = m.content.Update(msg)
	return m, cmd
}

func (m *Model) updateOptionState(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("UpdateOptionState: ", msg)
	switch msg := msg.(type) {
	case optionsMsg:
		m.options = msg.options
		m.loadOptionTable()
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.currentState = instanceState
			return m, tea.Batch(
				m.refreshInstances(),
				m.startTimer(m.refreshInterval, timerMsg{}),
			)
		case "enter":
			m.cursor = m.content.Cursor()
			m.detailsPanel = true
		case "r", "ctrl+l":
			return m, m.refreshOptions()
		}
	}

	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	return m, cmd
}

func (m *Model) updateFilesystemState(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("UpdateFilesystemState: ", msg)
	switch msg := msg.(type) {
	case filesystemsMsg:
		m.filesystems = msg.filesystems
		m.loadFilesystemTable()
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.currentState = sshState
			return m, nil
		case "enter":
			m.cursor = m.content.Cursor()
			m.detailsPanel = true
		}
	}

	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	return m, cmd
}

func (m *Model) updateSSHState(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("UpdateSSHState: ", msg)
	switch msg := msg.(type) {
	case sshKeysMsg:
		m.sshKeys = msg.sshKeys
		m.loadSSHTable()
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.currentState = instanceState
			return m, nil
		case "enter":
			m.cursor = m.content.Cursor()
			m.detailsPanel = true
		}
	}

	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	return m, cmd
}

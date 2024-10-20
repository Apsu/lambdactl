package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) launchCmd() tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.LaunchInstances(m.selectedOption, 1)
		if err != nil {
			return errMsg{err}
		}
		return m.refreshInstances()
	}
}

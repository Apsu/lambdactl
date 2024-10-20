package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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

package ui

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) startTimer(interval time.Duration, msg any) tea.Cmd {
	log.Println("Starting timer")
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return msg
	})
}

func (m *Model) refreshInstances() tea.Cmd {
	log.Println("Refreshing instances")
	return func() tea.Msg {
		instances, err := m.client.ListInstances()
		if err != nil {
			return errMsg{err}
		}

		return instancesMsg{instances}
	}
}

func (m *Model) refreshOptions() tea.Cmd {
	log.Println("Refreshing options")
	return func() tea.Msg {
		options, err := m.client.FetchInstanceOptions()
		if err != nil {
			return errMsg{err}
		}

		return optionsMsg{options}
	}
}

func (m *Model) refreshFilesystems() tea.Cmd {
	log.Println("Refreshing filesystems")
	return func() tea.Msg {
		filesystems, err := m.client.ListFilesystems()
		if err != nil {
			return errMsg{err}
		}

		return filesystemsMsg{filesystems}
	}
}

func (m *Model) refreshSSHKeys() tea.Cmd {
	log.Println("Refreshing SSH keys")
	return func() tea.Msg {
		sshKeys, err := m.client.ListSSHKeys()
		if err != nil {
			return errMsg{err}
		}

		return sshKeysMsg{sshKeys}
	}
}

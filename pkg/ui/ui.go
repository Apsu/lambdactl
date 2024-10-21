package ui

import (
	"fmt"
	"time"

	"lambdactl/pkg/api"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
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
		Cell: lipgloss.NewStyle().
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(draculaSelection)).
			Background(lipgloss.Color(draculaHighlight)),
	}

	contentTable := table.New(
		table.WithFocused(true),
		table.WithStyles(styles),
	)

	m := &Model{
		client:          client,
		refreshInterval: 30 * time.Second,
		errorTimeout:    5 * time.Second,
		currentState:    instanceState,
		header:          "Lambda Cloud",
		footer:          "Press 'q' to quit",
		content:         contentTable,
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshInstances(),
		m.startTimer(m.refreshInterval, timerMsg{}),
	)
}

func Start() error {
	program := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
	)

	_, err := program.Run()
	if err != nil {
		return fmt.Errorf("error running program: %v", err)
	}

	return nil
}

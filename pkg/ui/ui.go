package ui

import (
	"fmt"
	"time"

	"lambdactl/pkg/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

func NewModel() *Model {
	client := api.NewAPIClient(
		viper.GetString("api-url"),
		viper.GetString("api-key"),
	)

	m := &Model{
		client:          client,
		refreshInterval: 30 * time.Second,
		errorTimeout:    5 * time.Second,
		currentState:    runningState,
	}

	m.instanceTable = NewTable(
		TableRow{
			"Name",
			"Region",
			"Public IP",
			"Private IP",
			"Model",
			"GPUs",
			"CPUs",
			"RAM",
			"Storage",
			"Status",
		},
	)
	m.instanceTable.SetStyles(tableStyle, borderStyle, selectStyle)

	m.optionTable = NewTable(
		TableRow{
			"Region",
			"Model",
			"GPUs",
			"vCPUs",
			"Memory",
			"Storage",
			"$/Hour",
		},
	)
	m.optionTable.SetStyles(tableStyle, borderStyle, selectStyle)

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshInstances(),
		m.refreshOptions(),
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

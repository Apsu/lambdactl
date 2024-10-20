package ui

import (
	"fmt"
	"log"
	"os"
	"time"

	"lambdactl/pkg/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/spf13/viper"
)

func NewModel() *Model {
	client := api.NewAPIClient(
		viper.GetString("api-url"),
		viper.GetString("api-key"),
	)

	contentTable := table.New(
		table.WithFocused(true),
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
	// Create a log file
	if f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		log.Fatal(err)
	} else {
		defer f.Close()

		// Set log output to the file
		log.SetOutput(f)
	}

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

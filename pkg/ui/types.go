package ui

import (
	"lambdactl/pkg/api"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	runningState = "running"
	detailState  = "detail"
	optionState  = "option"
	launchState  = "launch"
	sshState     = "ssh"
)

const (
	draculaBackground  = "#282a36"
	draculaForeground  = "#f8f8f2"
	draculaError       = "#ff5555"
	draculaSelection   = "#44475a"
	draculaHeaderColor = "#bd93f9" // Purple for headers
	draculaHighlight   = "#ff79c6" // Pink for selected rows
)

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			BorderForeground(lipgloss.Color(draculaHeaderColor))
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(draculaError)).      // Dracula's red color
			Background(lipgloss.Color(draculaBackground)). // Dracula's background
			Padding(0, 1).
			Align(lipgloss.Center)
)

type Model struct {
	client          *api.APIClient
	machines        []api.InstanceDetails
	selectedMachine *api.InstanceDetails
	options         []api.InstanceOption
	selectedOption  *api.InstanceOption
	filter          string
	runningTable    table.Model
	optionTable     table.Model
	launchForm      huh.Form
	currentState    string
	previousState   string
	errorMsg        string
	refreshInterval time.Duration
	errorTimeout    time.Duration
}

type errMsg struct {
	err error
}

type clearErrMsg struct{}

func (e errMsg) Error() string {
	return e.err.Error()
}

type filterMsg struct {
	filter string
}

type instancesMsg struct {
	instances []api.InstanceDetails
}

type optionsMsg struct {
	options []api.InstanceOption
}

type timerMsg struct{}

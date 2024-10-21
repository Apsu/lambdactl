package ui

import (
	"time"

	"lambdactl/pkg/api"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh"
	lg "github.com/charmbracelet/lipgloss"
)

// Main state machine
const (
	instanceState AppState = iota
	optionState
	filesystemState
	sshState
)

// Theme colors
const (
	draculaBackground  = "#282a36"
	draculaForeground  = "#f8f8f2"
	draculaError       = "#ff5555" // Dracula red
	draculaSelection   = "#44475a" // Inverse foreground for selections
	draculaHeaderColor = "#bd93f9" // Purple for headers
	draculaHighlight   = "#ff79c6" // Pink background for selections
)

var ()

var (
	headerStyle = lg.NewStyle().
			Border(lg.RoundedBorder()).
			Background(lg.Color(draculaBackground)).
			Foreground(lg.Color(draculaHeaderColor)).
			Align(lg.Center)

	footerStyle = lg.NewStyle().
			BorderStyle(lg.RoundedBorder()).
			Background(lg.Color(draculaBackground)).
			Foreground(lg.Color(draculaHeaderColor))

	tableStyle = lg.NewStyle().
			Background(lg.Color(draculaBackground)).
			Foreground(lg.Color(draculaForeground))

	selectStyle = lg.NewStyle().
			Bold(true).
			Foreground(lg.Color(draculaSelection)).
			Background(lg.Color(draculaHighlight))

	modalStyle = lg.NewStyle().
			BorderStyle(lg.RoundedBorder()).
			Background(lg.Color(draculaBackground)).
			Padding(1, 2)

	borderStyle = lg.NewStyle().
			Foreground(lg.Color(draculaHeaderColor))

	errorStyle = lg.NewStyle().
			Align(lg.Center).
			Bold(true).
			Border(lg.RoundedBorder()).
			Foreground(lg.Color(draculaError)).
			Background(lg.Color(draculaBackground))
)

type AppState int

type Model struct {
	client *api.APIClient

	instances   []api.Instance
	options     []api.InstanceOption
	filesystems []api.Filesystem
	sshKeys     []api.SSHKey

	currentState AppState
	errorMsg     string
	filter       string

	header       string
	content      table.Model
	detailsPanel bool
	footer       string

	launchForm *huh.Form

	refreshInterval time.Duration
	errorTimeout    time.Duration

	// Terminal width/height
	width  int
	height int
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
	instances []api.Instance
}

type optionsMsg struct {
	options []api.InstanceOption
}

type filesystemsMsg struct {
	filesystems []api.Filesystem
}

type sshKeysMsg struct {
	sshKeys []api.SSHKey
}

type timerMsg struct{}

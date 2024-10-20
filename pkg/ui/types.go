package ui

import (
	"time"

	"lambdactl/pkg/api"

	"github.com/charmbracelet/huh"
	lg "github.com/charmbracelet/lipgloss"
	lgt "github.com/charmbracelet/lipgloss/table"
)

const (
	runningState = "running"
	detailState  = "detail"
	optionState  = "options"
	launchState  = "launch"
	sshState     = "ssh"
)

const (
	draculaBackground  = "#282a36"
	draculaForeground  = "#f8f8f2"
	draculaError       = "#ff5555" // Dracula red
	draculaSelection   = "#44475a" // Inverse foreground for selections
	draculaHeaderColor = "#bd93f9" // Purple for headers
	draculaHighlight   = "#ff79c6" // Pink background for selections
)

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

type Model struct {
	client *api.APIClient

	instances []api.Instance
	options   []api.InstanceOption

	selectedInstance api.Instance
	selectedOption   api.InstanceOption

	currentState string
	errorMsg     string
	filter       string

	header        string
	instanceTable *Table
	optionTable   *Table
	footer        string

	launchForm *huh.Form

	refreshInterval time.Duration
	errorTimeout    time.Duration

	// Terminal width/height
	width  int
	height int
}

type Table struct {
	table *lgt.Table

	borderStyle lg.Style
	selectStyle lg.Style
	tableStyle  lg.Style

	headers TableRow
	rows    TableRows

	length int

	width    int // Total table width
	height   int // Total table height
	viewport int // Content viewport height

	cursor int // 1 = first row
	offset int // 0 = no row shift
}

type TableRow []string
type TableRows []TableRow

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

type timerMsg struct{}

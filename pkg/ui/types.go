package ui

import (
	"lambdactl/pkg/api"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

type State struct {
	Name string
	View func(Model) string
}

var (
	listState   = State{Name: "list", View: listView}
	detailState = State{Name: "detail", View: detailView}
	launchState = State{Name: "launch", View: launchView}
)

type Model struct {
	client          *api.APIClient
	machines        []api.InstanceDetails
	selectedMachine *api.InstanceDetails
	filter          string
	table           table.Model
	currentState    State
	previousState   State
	errorMsg        string
	refreshInterval time.Duration
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type filterMsg struct {
	filter string
}

type createdVMMsg struct{}

type instancesMsg struct {
	instances []api.InstanceDetails
}

type timerMsg struct{}

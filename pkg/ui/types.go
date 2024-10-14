package ui

import (
	"lambdactl/pkg/api"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

const (
	runningState = "running"
	detailState  = "detail"
	optionState  = "option"
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
	currentState    string
	previousState   string
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

type optionsMsg struct {
	options []api.InstanceOption
}

type timerMsg struct{}

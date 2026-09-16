package arguments

import (
	"errors"
	"strings"

	"github.com/unhealme/lakehouse-admin-tools/internal/clients/yarn"
)

type YarnListAppsArgs struct {
	Queue      string                `arg:"-Q,--queue" help:"filter applications by queue" placeholder:"NAME"`
	Format     YarnListAppsOutputFmt `arg:"-f,--format" default:"csv" help:"output format" placeholder:"{csv,json}"`
	Limit      int                   `arg:"-l,--limit" help:"applications limit" placeholder:"NUM"`
	States     yarnListAppsStates    `arg:"-s,--states" help:"yarn application states as a comma delimited string" placeholder:"STATES"`
	User       string                `arg:"-u,--user" help:"filter applications by user" placeholder:"NAME"`
	OutputFile string                `arg:"-o,--" help:"write result to FILE instead of stdout" placeholder:"FILE"`
	NoHeader   bool                  `arg:"-,--no-header" help:"do not print header for csv output format"`

	YarnClient *yarn.YarnRMClient `arg:"-"`
}

type yarnListAppsStates []yarn.ApplicationState

func (f *yarnListAppsStates) UnmarshalText(buf []byte) error {
	var states yarnListAppsStates
	for state := range strings.SplitSeq(string(buf), ",") {
		var appState yarn.ApplicationState
		switch strings.ToUpper(strings.TrimSpace(state)) {
		case "NEW":
			appState = yarn.NEW
		case "NEW_SAVING":
			appState = yarn.NEW_SAVING
		case "SUBMITTED":
			appState = yarn.SUBMITTED
		case "ACCEPTED":
			appState = yarn.ACCEPTED
		case "RUNNING":
			appState = yarn.RUNNING
		case "FINISHED":
			appState = yarn.FINISHED
		case "FAILED":
			appState = yarn.FAILED
		case "KILLED":
			appState = yarn.KILLED
		default:
			return errors.New("invalid application state: " + state)
		}
		states = append(states, appState)
	}
	*f = states
	return nil
}

type YarnListAppsOutputFmt int

const (
	YarnListAppsOutputCsv YarnListAppsOutputFmt = iota + 1
	YarnListAppsOutputJson
)

func (f *YarnListAppsOutputFmt) UnmarshalText(buf []byte) error {
	switch fmt := string(buf); strings.ToLower(strings.TrimSpace(fmt)) {
	case "csv":
		*f = YarnListAppsOutputCsv
	case "json":
		*f = YarnListAppsOutputJson
	default:
		return errors.New("invalid output format: " + fmt)
	}
	return nil
}

package tui

import (
	"context"

	"github.com/elisiariocouto/container-toolbox/internal/stacks"
)

// StackStatus is a discovered stack joined with its live container status.
type StackStatus struct {
	stacks.Stack
	Running       int
	Total         int
	HasContainers bool // true if any containers exist for this project
}

// State of the stack relative to its containers.
type runState int

const (
	stateUnknown runState = iota // status could not be determined (daemon down)
	stateStopped                 // no running containers
	statePartial                 // some but not all containers running
	stateRunning                 // all containers running
)

func (s StackStatus) state(daemonOK bool) runState {
	if !daemonOK {
		return stateUnknown
	}
	switch {
	case s.Total == 0 || s.Running == 0:
		return stateStopped
	case s.Running == s.Total:
		return stateRunning
	default:
		return statePartial
	}
}

// --- custom messages ---

type stacksLoadedMsg struct {
	stacks    []StackStatus
	daemonOK  bool
	daemonErr error
}

type actionDoneMsg struct {
	stack  string
	action string
	err    error
}

type logsStartedMsg struct {
	stack  string
	sub    chan string
	cancel context.CancelFunc
}

type logLineMsg struct{ line string }

type logsClosedMsg struct{ err error }

type errMsg struct{ err error }

package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// stackItem adapts a StackStatus to the bubbles list.Item interface.
type stackItem struct {
	status   StackStatus
	daemonOK bool
}

func (i stackItem) Title() string { return i.status.Name }

func (i stackItem) Description() string {
	st := i.status.state(i.daemonOK)
	dot := lipgloss.NewStyle().Foreground(stateColor(st))
	switch st {
	case stateRunning:
		return dot.Render("●") + fmt.Sprintf(" running %d/%d", i.status.Running, i.status.Total)
	case statePartial:
		return dot.Render("●") + fmt.Sprintf(" partial %d/%d", i.status.Running, i.status.Total)
	case stateUnknown:
		return dot.Render("●") + " unknown"
	default:
		return dot.Render("○") + " stopped"
	}
}

// FilterValue lets the list filter by stack name.
func (i stackItem) FilterValue() string { return i.status.Name }

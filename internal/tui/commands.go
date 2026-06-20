package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/elisiariocouto/container-toolbox/internal/docker"
	"github.com/elisiariocouto/container-toolbox/internal/stacks"
)

const actionTimeout = 5 * time.Minute

// loadStacksCmd discovers stacks on disk and joins them with live container
// status from the daemon. A daemon error is reported but does not prevent the
// discovered stacks from being shown (with unknown status).
func loadStacksCmd(dir string, dc *docker.Client) tea.Cmd {
	return func() tea.Msg {
		found, err := stacks.Discover(dir)
		if err != nil {
			return errMsg{err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		byProject, derr := dc.ListContainersByProject(ctx)
		if derr != nil {
			return stacksLoadedMsg{stacks: join(found, nil), daemonOK: false, daemonErr: derr}
		}
		return stacksLoadedMsg{stacks: join(found, byProject), daemonOK: true}
	}
}

// join layers live status (by Compose project name, with a working_dir
// fallback) onto the filesystem-discovered stacks.
func join(found []stacks.Stack, byProject map[string]docker.ProjectStatus) []StackStatus {
	res := make([]StackStatus, 0, len(found))
	for _, s := range found {
		ps, ok := byProject[s.Project]
		if !ok {
			// fallback: match on the working_dir label
			for _, candidate := range byProject {
				if candidate.WorkingDir == s.Dir {
					ps, ok = candidate, true
					break
				}
			}
		}
		res = append(res, StackStatus{
			Stack:         s,
			Running:       ps.Running,
			Total:         ps.Total,
			HasContainers: ok,
		})
	}
	return res
}

// runComposeCmd executes a lifecycle action and reports completion.
func runComposeCmd(s stacks.Stack, action string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), actionTimeout)
		defer cancel()

		var err error
		switch action {
		case "up":
			err = docker.Up(ctx, s.Dir)
		case "down":
			err = docker.Down(ctx, s.Dir)
		case "restart":
			err = docker.Restart(ctx, s.Dir)
		case "pull":
			err = docker.PullUp(ctx, s.Dir)
		}
		return actionDoneMsg{stack: s.Name, action: action, err: err}
	}
}

// startLogsCmd begins streaming logs for a stack into a fresh channel.
func startLogsCmd(s stacks.Stack) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan string, 256)
		if err := docker.StreamLogs(ctx, s.Dir, ch); err != nil {
			cancel()
			return logsClosedMsg{err}
		}
		return logsStartedMsg{stack: s.Name, sub: ch, cancel: cancel}
	}
}

// waitForLogCmd reads one line from the channel and reschedules itself, turning
// a streaming channel into a sequence of Bubble Tea messages.
func waitForLogCmd(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return logsClosedMsg{}
		}
		return logLineMsg{line}
	}
}

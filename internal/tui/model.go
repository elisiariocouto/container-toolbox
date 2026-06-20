package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/elisiariocouto/container-toolbox/internal/docker"
)

type viewState int

const (
	stateList viewState = iota
	stateLogs
)

// Model is the root Bubble Tea model.
type Model struct {
	state     viewState
	stacksDir string
	docker    *docker.Client

	list     list.Model
	viewport viewport.Model
	spinner  spinner.Model
	help     help.Model

	daemonOK  bool
	daemonErr error

	busy      bool
	busyLabel string
	statusMsg string
	err       error

	// log streaming
	logSub    chan string
	logCancel context.CancelFunc
	logBuf    *strings.Builder
	logStack  string

	width, height int
}

// New constructs the root model.
func New(stacksDir string, dc *docker.Client) Model {
	delegate := list.NewDefaultDelegate()

	l := list.New(nil, delegate, 0, 0)
	l.Title = "container-toolbox"
	l.Styles.Title = titleStyle
	l.SetShowHelp(false) // we render our own help bar
	l.SetStatusBarItemName("stack", "stacks")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = busyStyle

	return Model{
		state:     stateList,
		stacksDir: stacksDir,
		docker:    dc,
		list:      l,
		viewport:  viewport.New(0, 0),
		spinner:   sp,
		help:      help.New(),
		daemonOK:  true,
		logBuf:    &strings.Builder{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadStacksCmd(m.stacksDir, m.docker), m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case stacksLoadedMsg:
		m.daemonOK = msg.daemonOK
		m.daemonErr = msg.daemonErr
		items := make([]list.Item, 0, len(msg.stacks))
		for _, s := range msg.stacks {
			items = append(items, stackItem{status: s, daemonOK: msg.daemonOK})
		}
		cmd := m.list.SetItems(items)
		return m, cmd

	case actionDoneMsg:
		m.busy = false
		m.busyLabel = ""
		if msg.err != nil {
			m.statusMsg = errorStyle.Render(fmt.Sprintf("%s %s failed: %s", msg.action, msg.stack, firstLine(msg.err.Error())))
		} else {
			m.statusMsg = fmt.Sprintf("%s %s ✓", msg.action, msg.stack)
		}
		// refresh status after any action
		return m, loadStacksCmd(m.stacksDir, m.docker)

	case logsStartedMsg:
		m.state = stateLogs
		m.logSub = msg.sub
		m.logCancel = msg.cancel
		m.logStack = msg.stack
		m.logBuf.Reset()
		m.viewport.SetContent("")
		m.layout()
		return m, waitForLogCmd(m.logSub)

	case logLineMsg:
		m.logBuf.WriteString(msg.line)
		m.logBuf.WriteByte('\n')
		m.viewport.SetContent(m.logBuf.String())
		m.viewport.GotoBottom()
		return m, waitForLogCmd(m.logSub)

	case logsClosedMsg:
		if msg.err != nil {
			m.statusMsg = errorStyle.Render("logs: " + firstLine(msg.err.Error()))
		}
		m.stopLogs()
		m.state = stateList
		m.layout()
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// delegate remaining messages to the active component
	return m.updateActive(msg)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When filtering in the list, let it consume keystrokes.
	if m.state == stateList && m.list.FilterState() == list.Filtering {
		return m.updateActive(msg)
	}

	switch m.state {
	case stateLogs:
		switch {
		case key.Matches(msg, keys.Back):
			m.stopLogs()
			m.state = stateList
			return m, nil
		case key.Matches(msg, keys.Quit):
			return m, m.quit()
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case stateList:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, m.quit()
		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.layout()
			return m, nil
		case key.Matches(msg, keys.Refresh):
			m.statusMsg = "refreshing…"
			return m, loadStacksCmd(m.stacksDir, m.docker)
		case key.Matches(msg, keys.Logs):
			if s, ok := m.selected(); ok {
				return m, startLogsCmd(s.Stack)
			}
			return m, nil
		case key.Matches(msg, keys.Up):
			return m, m.dispatch("up", "Starting")
		case key.Matches(msg, keys.Down):
			return m, m.dispatch("down", "Stopping")
		case key.Matches(msg, keys.Restart):
			return m, m.dispatch("restart", "Restarting")
		case key.Matches(msg, keys.Pull):
			return m, m.dispatch("pull", "Updating")
		}
	}

	return m.updateActive(msg)
}

// dispatch starts a lifecycle action on the selected stack, unless one is
// already running.
func (m *Model) dispatch(action, verb string) tea.Cmd {
	if m.busy {
		return nil
	}
	s, ok := m.selected()
	if !ok {
		return nil
	}
	m.busy = true
	m.busyLabel = fmt.Sprintf("%s %s…", verb, s.Name)
	m.statusMsg = ""
	return tea.Batch(runComposeCmd(s.Stack, action), m.spinner.Tick)
}

func (m Model) updateActive(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case stateLogs:
		m.viewport, cmd = m.viewport.Update(msg)
	default:
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.state {
	case stateLogs:
		return m.logsView()
	default:
		return m.listView()
	}
}

func (m Model) listView() string {
	if len(m.list.Items()) == 0 {
		return m.chrome(emptyStyle.Render(fmt.Sprintf("No stacks found in %s", m.stacksDir)))
	}
	return m.chrome(m.list.View())
}

// chrome wraps body with the footer (status/help bars).
func (m Model) chrome(body string) string {
	return lipgloss.JoinVertical(lipgloss.Left, body, m.footer())
}

func (m Model) footer() string {
	var lines []string

	if m.busy {
		lines = append(lines, busyStyle.Render(m.spinner.View()+m.busyLabel))
	} else if m.statusMsg != "" {
		lines = append(lines, statusBarStyle.Render(m.statusMsg))
	}

	if !m.daemonOK {
		msg := "Docker daemon unreachable — is it running?"
		if m.daemonErr != nil {
			msg = "Docker daemon: " + firstLine(m.daemonErr.Error())
		}
		lines = append(lines, errorStyle.Render(msg))
	}
	if m.err != nil {
		lines = append(lines, errorStyle.Render(firstLine(m.err.Error())))
	}

	lines = append(lines, statusBarStyle.Render(m.help.View(keys)))
	return strings.Join(lines, "\n")
}

func (m Model) logsView() string {
	title := logsTitleStyle.Render("logs: " + m.logStack)
	help := statusBarStyle.Render("esc/q back · ↑/↓ scroll · ctrl+c quit")
	return lipgloss.JoinVertical(lipgloss.Left, title, viewportStyle.Render(m.viewport.View()), help)
}

// layout recomputes component sizes from the current terminal size.
func (m *Model) layout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	footerHeight := lipgloss.Height(m.footer())
	m.list.SetSize(m.width, max(m.height-footerHeight, 1))

	// viewport: leave room for title (1), border (2), help (1)
	vpHeight := max(m.height-4, 1)
	vpWidth := max(m.width-2, 1)
	m.viewport.Width = vpWidth
	m.viewport.Height = vpHeight
}

func (m *Model) stopLogs() {
	if m.logCancel != nil {
		m.logCancel()
		m.logCancel = nil
	}
	m.logSub = nil
}

func (m Model) quit() tea.Cmd {
	if m.logCancel != nil {
		m.logCancel()
	}
	return tea.Quit
}

func (m Model) selected() (StackStatus, bool) {
	it, ok := m.list.SelectedItem().(stackItem)
	if !ok {
		return StackStatus{}, false
	}
	return it.status, true
}

// --- helpers ---

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

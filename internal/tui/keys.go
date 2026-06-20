package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap holds every binding. Movement keys are handled by the list/viewport
// components natively; the bindings here are the app-level actions.
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Restart key.Binding
	Pull    key.Binding
	Logs    key.Binding
	Refresh key.Binding
	Back    key.Binding
	Help    key.Binding
	Quit    key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "up -d"),
	),
	Down: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "down"),
	),
	Restart: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "restart"),
	),
	Pull: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pull+up"),
	),
	Logs: key.NewBinding(
		key.WithKeys("enter", "l"),
		key.WithHelp("enter/l", "logs"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "refresh"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc", "back"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// ShortHelp is shown in the list view help bar.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Restart, k.Pull, k.Logs, k.Refresh, k.Help, k.Quit}
}

// FullHelp is shown when help is expanded.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Restart, k.Pull},
		{k.Logs, k.Refresh, k.Back, k.Help, k.Quit},
	}
}

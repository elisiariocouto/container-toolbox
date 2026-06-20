package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorRunning = lipgloss.Color("42")  // green
	colorStopped = lipgloss.Color("244") // gray
	colorPartial = lipgloss.Color("214") // amber
	colorUnknown = lipgloss.Color("203") // red
	colorAccent  = lipgloss.Color("63")  // purple
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(colorAccent).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorUnknown).
			Padding(0, 1)

	busyStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Padding(0, 1)

	logsTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(colorAccent).
			Padding(0, 1)

	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent)

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 2)
)

// stateColor maps a run state to its display color.
func stateColor(s runState) lipgloss.Color {
	switch s {
	case stateRunning:
		return colorRunning
	case statePartial:
		return colorPartial
	case stateUnknown:
		return colorUnknown
	default:
		return colorStopped
	}
}

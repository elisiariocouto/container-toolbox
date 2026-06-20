// Command container-toolbox is a terminal UI for managing Docker Compose stacks
// that live as subdirectories of ~/stacks on the local machine.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/elisiariocouto/container-toolbox/internal/docker"
	"github.com/elisiariocouto/container-toolbox/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	defaultDir := filepath.Join(homeDir(), "stacks")
	stacksDir := flag.String("stacks-dir", defaultDir, "directory containing compose stacks")
	flag.Parse()

	if err := docker.ComposeAvailable(); err != nil {
		return err
	}

	dc, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("docker client: %w", err)
	}
	defer dc.Close()

	model := tui.New(*stacksDir, dc)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}

package tui

import (
	"testing"

	"github.com/elisiariocouto/container-toolbox/internal/docker"
	"github.com/elisiariocouto/container-toolbox/internal/stacks"
)

func TestJoin(t *testing.T) {
	found := []stacks.Stack{
		{Name: "web", Dir: "/stacks/web", Project: "web"},
		{Name: "db", Dir: "/stacks/db", Project: "db"},
		{Name: "My App", Dir: "/stacks/My App", Project: "myapp"},
		{Name: "idle", Dir: "/stacks/idle", Project: "idle"},
	}
	byProject := map[string]docker.ProjectStatus{
		"web": {Running: 2, Total: 2},
		"db":  {Running: 1, Total: 3},
		// "My App" sanitizes to "myapp" but suppose the label drifted; match by working_dir.
		"someproj": {Running: 1, Total: 1, WorkingDir: "/stacks/My App"},
		// "idle" has no containers at all.
	}

	got := join(found, byProject)
	if len(got) != 4 {
		t.Fatalf("got %d, want 4", len(got))
	}

	byName := map[string]StackStatus{}
	for _, s := range got {
		byName[s.Name] = s
	}

	if s := byName["web"]; !s.HasContainers || s.Running != 2 || s.Total != 2 {
		t.Errorf("web: %+v", s)
	}
	if s := byName["db"]; s.state(true) != statePartial {
		t.Errorf("db should be partial, got state for %+v", s)
	}
	if s := byName["My App"]; !s.HasContainers || s.Running != 1 {
		t.Errorf("My App should match via working_dir fallback: %+v", s)
	}
	if s := byName["idle"]; s.HasContainers || s.state(true) != stateStopped {
		t.Errorf("idle should be stopped with no containers: %+v", s)
	}
	if s := byName["web"]; s.state(false) != stateUnknown {
		t.Errorf("daemon down should yield unknown: %+v", s)
	}
}

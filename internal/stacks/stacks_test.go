package stacks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectName(t *testing.T) {
	cases := map[string]string{
		"web":       "web",
		"My App":    "myapp",
		"Foo_Bar-1": "foo_bar-1",
		"_hidden":   "hidden",
		"123abc":    "123abc",
		"a.b.c":     "abc",
		"--lead":    "lead",
	}
	for in, want := range cases {
		if got := ProjectName(in); got != want {
			t.Errorf("ProjectName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDiscover(t *testing.T) {
	root := t.TempDir()

	// stack with compose.yaml
	mkStack(t, root, "alpha", "compose.yaml")
	// stack with docker-compose.yml
	mkStack(t, root, "beta", "docker-compose.yml")
	// directory without a compose file -> ignored
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	// a regular file at the root -> ignored
	if err := os.WriteFile(filepath.Join(root, "README"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d stacks, want 2: %+v", len(got), got)
	}
	if got[0].Name != "alpha" || got[1].Name != "beta" {
		t.Fatalf("unexpected names/order: %+v", got)
	}
	if got[0].Project != "alpha" {
		t.Errorf("project = %q, want alpha", got[0].Project)
	}
	if filepath.Base(got[1].ComposeFile) != "docker-compose.yml" {
		t.Errorf("compose file = %q", got[1].ComposeFile)
	}
}

func TestDiscoverMissingRoot(t *testing.T) {
	got, err := Discover(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("expected no error for missing root, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %+v", got)
	}
}

func mkStack(t *testing.T, root, name, composeFile string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, composeFile), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

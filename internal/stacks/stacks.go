// Package stacks discovers Docker Compose stacks on the local filesystem.
//
// A stack is a direct subdirectory of the stacks root (default ~/stacks) that
// contains a compose file. The directory name is the stack name, and the
// Compose project name is derived from it using Compose's own sanitization
// rules so it can be joined against the labels Compose stamps on containers.
package stacks

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// composeFilenames are the file names Compose recognizes, in preference order.
var composeFilenames = []string{
	"compose.yaml",
	"compose.yml",
	"docker-compose.yaml",
	"docker-compose.yml",
}

// Stack is a compose project discovered on disk.
type Stack struct {
	Name        string // directory name, shown to the user
	Dir         string // absolute path to the stack directory
	ComposeFile string // absolute path to the detected compose file
	Project     string // sanitized Compose project name (join key for labels)
}

// Discover returns every stack found directly under root, sorted by name.
//
// A missing root is not an error: it returns an empty slice so the UI can show
// a friendly empty state. Real I/O errors (e.g. permission denied) are returned.
func Discover(root string) ([]Stack, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var found []Stack
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		composeFile, ok := findComposeFile(dir)
		if !ok {
			continue
		}
		found = append(found, Stack{
			Name:        e.Name(),
			Dir:         dir,
			ComposeFile: composeFile,
			Project:     ProjectName(e.Name()),
		})
	}

	sort.Slice(found, func(i, j int) bool { return found[i].Name < found[j].Name })
	return found, nil
}

// findComposeFile returns the path to the first recognized compose file in dir.
func findComposeFile(dir string) (string, bool) {
	for _, name := range composeFilenames {
		p := filepath.Join(dir, name)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, true
		}
	}
	return "", false
}

// ProjectName derives the Compose project name from a directory name using the
// same normalization Compose applies: lowercase, keep only [a-z0-9_-], and
// strip leading characters that are not a letter or digit.
func ProjectName(dir string) string {
	lower := strings.ToLower(dir)
	var b strings.Builder
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		default:
			// drop everything else (spaces, dots, etc.)
		}
	}
	out := b.String()
	out = strings.TrimLeft(out, "_-")
	return out
}

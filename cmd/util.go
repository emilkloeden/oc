package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func projectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return findProjectRoot(cwd)
}

func findProjectRoot(start string) (string, error) {
	dir := start
	for {
		if isProjectDir(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no OCaml project found (no dune-project or .opam file in %s or any parent directory)", start)
}

// isProjectDir reports whether dir looks like an OCaml project root.
// Recognised markers: dune-project, *.opam file, or .oc/ directory (oc state).
func isProjectDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "dune-project")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, ".oc")); err == nil {
		return true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".opam") {
			return true
		}
	}
	return false
}

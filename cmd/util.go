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

// isProjectDir reports whether dir looks like an OCaml project root that oc can manage.
// Recognised markers (in priority order):
//  1. dune-project with (generate_opam_files — dune-managed project
//  2. *.opam file — hand-written opam project
//  3. .oc/ directory — machine-local oc state (project already initialised)
//
// A bare dune-project without generate_opam_files and without a .opam file
// is not accepted, keeping this consistent with project.Detect.
func isProjectDir(dir string) bool {
	if data, err := os.ReadFile(filepath.Join(dir, "dune-project")); err == nil {
		if strings.Contains(string(data), "(generate_opam_files") {
			return true
		}
	}
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".opam") {
				return true
			}
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".oc")); err == nil {
		return true
	}
	return false
}

package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectType int

const (
	TypeDuneManaged     ProjectType = iota
	TypeHandWrittenOpam ProjectType = iota
)

// Detect determines the project type for dir.
// Dune-managed (dune-project with generate_opam_files) takes priority.
// Falls back to a hand-written .opam file.
// Returns an error if no recognized project format is found.
func Detect(dir string) (ProjectType, error) {
	dpPath := filepath.Join(dir, "dune-project")
	if data, err := os.ReadFile(dpPath); err == nil {
		if strings.Contains(string(data), "(generate_opam_files") {
			return TypeDuneManaged, nil
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read directory: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".opam") {
			return TypeHandWrittenOpam, nil
		}
	}

	return 0, fmt.Errorf("no OCaml project found in %s (no dune-project with generate_opam_files, no .opam file)", dir)
}

package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emilkloeden/oc/cmd"
)

func TestRunRemove_OpamFailurePropagates(t *testing.T) {
	dir := t.TempDir()

	// Write a dune-project with fakedep in depends.
	duneProject := "(lang dune 3.0)\n(generate_opam_files true)\n\n(package\n (name myapp)\n (depends\n  (ocaml (>= \"5.2.0\"))\n  dune\n  fakedep))\n"
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte(duneProject), 0644); err != nil {
		t.Fatal(err)
	}

	// RunRemove should propagate the opam error instead of silently ignoring it.
	// Since opam is either not installed or "fakedep" is not a real package,
	// the opam remove call will fail and the error must be returned.
	err := cmd.RunRemove(dir, "fakedep")
	if err == nil {
		t.Fatal("expected error when opam remove fails, got nil")
	}
}

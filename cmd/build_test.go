package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emilkloeden/oc/cmd"
)

func TestRunBuild_CorruptedLock_ReturnsError(t *testing.T) {
	dir := t.TempDir()

	// Write a valid dune-project so findProjectRoot succeeds.
	duneProject := "(lang dune 3.0)\n(generate_opam_files true)\n\n(package\n (name myapp)\n (depends\n  (ocaml (>= \"5.2.0\"))\n  dune))\n"
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte(duneProject), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a corrupted oc.lock file.
	if err := os.WriteFile(filepath.Join(dir, "oc.lock"), []byte("NOT VALID TOML ][[["), 0644); err != nil {
		t.Fatal(err)
	}

	// RunBuild should return an error (not panic) when oc.lock is corrupted.
	err := cmd.RunBuild(dir)
	if err == nil {
		t.Fatal("expected error when oc.lock is corrupted, got nil")
	}
}

package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emilkloeden/oc/cmd"
)

func TestRunRemove_OpamFailurePropagates(t *testing.T) {
	dir := t.TempDir()

	// Write a valid oc.toml with a dependency.
	content := "[project]\nname = \"myapp\"\nversion = \"0.1.0\"\n\n[ocaml]\nversion = \"5.2.0\"\n\n[dependencies]\nfakedep = \"*\"\n"
	if err := os.WriteFile(filepath.Join(dir, "oc.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a stub .opam file so opam.Generate doesn't need to create it.
	opamContent := "opam-version: \"2.0\"\nname: \"myapp\"\n"
	if err := os.WriteFile(filepath.Join(dir, "myapp.opam"), []byte(opamContent), 0644); err != nil {
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

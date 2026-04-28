package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emilkloeden/oc/cmd"
)

// validOcToml returns minimal valid oc.toml content for the given project name.
func validOcToml(name string) string {
	return "[project]\nname = \"" + name + "\"\nversion = \"0.1.0\"\n\n[ocaml]\nversion = \"5.2.0\"\n"
}

func TestRunBuild_CorruptedLock_ReturnsError(t *testing.T) {
	dir := t.TempDir()

	// Write a valid oc.toml so LoadConfig succeeds.
	if err := os.WriteFile(filepath.Join(dir, "oc.toml"), []byte(validOcToml("myapp")), 0644); err != nil {
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

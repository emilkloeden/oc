package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emilkloeden/oc/cmd"
	"github.com/emilkloeden/oc/internal/project"
)

const oldLockContent = `
switch_path = "/home/alice/.cache/oc/switches/abc123def456abcd"

[ocaml]
  version = "5.2.0"

[[package]]
  name = "dune"
  version = "3.15.0"
`

func TestMigrateOcLock_MigratesWhenLockExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "oc.lock"), []byte(oldLockContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd.MigrateIfNeeded(dir)

	// oc.lock should be deleted
	if _, err := os.Stat(filepath.Join(dir, "oc.lock")); !os.IsNotExist(err) {
		t.Error("expected oc.lock to be deleted after migration")
	}

	// .oc/state.toml should exist with the switch_path
	s, err := project.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState after migration: %v", err)
	}
	if s.SwitchPath != "/home/alice/.cache/oc/switches/abc123def456abcd" {
		t.Errorf("SwitchPath after migration: got %q", s.SwitchPath)
	}
}

func TestMigrateOcLock_NoopWhenNoLock(t *testing.T) {
	dir := t.TempDir()
	// Should not error or create any files
	cmd.MigrateIfNeeded(dir)
	if _, err := os.Stat(filepath.Join(dir, ".oc", "state.toml")); !os.IsNotExist(err) {
		t.Error("expected no state.toml to be created when there is no oc.lock")
	}
}

func TestMigrateOcLock_NoopWhenStateAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	// Both files exist — state is newer, lock is leftover
	if err := os.WriteFile(filepath.Join(dir, "oc.lock"), []byte(oldLockContent), 0644); err != nil {
		t.Fatal(err)
	}
	existing := project.State{SwitchPath: "/existing/path", OCamlVersion: "5.3.0"}
	if err := project.SaveState(dir, existing); err != nil {
		t.Fatal(err)
	}

	cmd.MigrateIfNeeded(dir)

	// State should be unchanged (not overwritten by migration)
	s, err := project.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if s.SwitchPath != "/existing/path" {
		t.Errorf("migration should not overwrite existing state; got SwitchPath %q", s.SwitchPath)
	}
}

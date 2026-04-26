//go:build integration

package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/cmd"
	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
)

func requireOpam(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("opam"); err != nil {
		t.Skip("opam not in PATH")
	}
}

// TestIntegration_NewCreatesValidProject verifies oc new produces a project
// that opam and dune can understand.
func TestIntegration_NewCreatesValidProject(t *testing.T) {
	requireOpam(t)

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	projectDir := filepath.Join(dir, "hello")

	// opam lint should accept the generated .opam file
	out, err := runCmd(projectDir, "opam", "lint", "hello.opam")
	if err != nil {
		t.Fatalf("opam lint failed: %v\noutput: %s", err, out)
	}
}

// TestIntegration_SyncCreatesSwitch verifies EnsureWith (real runner) creates
// a functional opam switch.
func TestIntegration_SyncCreatesSwitch(t *testing.T) {
	requireOpam(t)
	t.Log("This test creates an opam switch — may take several minutes on first run.")

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}
	projectDir := filepath.Join(dir, "hello")

	cfg, err := project.LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if err := sync.Ensure(projectDir, cfg); err != nil {
		t.Fatalf("sync.Ensure: %v", err)
	}

	// .ocaml symlink must exist
	link := filepath.Join(projectDir, ".ocaml")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf(".ocaml not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error(".ocaml should be a symlink")
	}

	// lockfile must be written
	lock, err := project.LoadLock(projectDir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if lock.OCaml.Version == "" {
		t.Error("lockfile missing OCaml version")
	}
}

// TestIntegration_SwitchPathStoredInLock verifies that after sync the switch
// path is persisted in oc.lock so subsequent calls stay on the same switch.
func TestIntegration_SwitchPathStoredInLock(t *testing.T) {
	requireOpam(t)
	t.Log("Creating a switch to verify lock persistence — may take several minutes on first run.")

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}
	projectDir := filepath.Join(dir, "hello")

	cfg, err := project.LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if err := sync.Ensure(projectDir, cfg); err != nil {
		t.Fatalf("first sync.Ensure: %v", err)
	}

	lock1, _ := project.LoadLock(projectDir)
	path1 := lock1.SwitchPath

	// Second sync — lock now has packages; without storing path, hash would differ.
	if err := sync.Ensure(projectDir, cfg); err != nil {
		t.Fatalf("second sync.Ensure: %v", err)
	}

	lock2, _ := project.LoadLock(projectDir)
	path2 := lock2.SwitchPath

	if path1 == "" {
		t.Fatal("SwitchPath not stored after first sync")
	}
	if path1 != path2 {
		t.Errorf("switch path changed between syncs: %q → %q", path1, path2)
	}
}

// TestIntegration_BuildHelloWorld verifies the full oc new → sync → dune build
// pipeline produces a working executable.
func TestIntegration_BuildHelloWorld(t *testing.T) {
	requireOpam(t)
	t.Log("This test performs a full build — may take several minutes on first run.")

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}
	projectDir := filepath.Join(dir, "hello")

	cfg, err := project.LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if err := sync.Ensure(projectDir, cfg); err != nil {
		t.Fatalf("sync.Ensure: %v", err)
	}

	lock, err := project.LoadLock(projectDir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	switchPath := lock.SwitchPath
	if switchPath == "" {
		t.Fatal("lock.SwitchPath empty after sync")
	}

	out, err := runCmd(projectDir, "opam", "exec", "--switch", switchPath, "--", "dune", "build")
	if err != nil {
		t.Fatalf("dune build failed: %v\noutput: %s", err, out)
	}

	exe := filepath.Join(projectDir, "_build", "default", "bin", "main.exe")
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("expected built executable at %s: %v", exe, err)
	}
}

// TestIntegration_AddDependency verifies oc add updates the lockfile.
func TestIntegration_AddDependency(t *testing.T) {
	requireOpam(t)
	t.Log("This test installs an opam package — may take several minutes.")

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}
	projectDir := filepath.Join(dir, "hello")

	cfg, err := project.LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// Add a small dep with no C deps for speed.
	cfg.Dependencies["stringext"] = "*"
	if err := project.SaveConfig(projectDir, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if err := opam.Generate(projectDir, cfg); err != nil {
		t.Fatalf("opam.Generate: %v", err)
	}

	if err := sync.Ensure(projectDir, cfg); err != nil {
		t.Fatalf("sync.Ensure: %v", err)
	}

	lock, err := project.LoadLock(projectDir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}

	found := false
	for _, p := range lock.Packages {
		if strings.EqualFold(p.Name, "stringext") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("stringext not found in lockfile packages: %v", lock.Packages)
	}
}

func runCmd(dir, name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}

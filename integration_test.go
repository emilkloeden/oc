//go:build integration

package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/cmd"
	"github.com/emilkloeden/oc/internal/dune"
	ocExec "github.com/emilkloeden/oc/internal/exec"
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
// that opam and dune can understand (via dune generating the .opam file).
func TestIntegration_NewCreatesValidProject(t *testing.T) {
	requireOpam(t)

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	projectDir := filepath.Join(dir, "hello")

	// dune-project should have generate_opam_files
	if !dune.HasGenerateOpamFiles(projectDir) {
		t.Error("dune-project missing (generate_opam_files true)")
	}
}

// TestIntegration_SyncCreatesSwitch verifies sync.Ensure (real runner) creates
// a functional opam switch.
func TestIntegration_SyncCreatesSwitch(t *testing.T) {
	requireOpam(t)
	t.Log("This test creates an opam switch — may take several minutes on first run.")

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}
	projectDir := filepath.Join(dir, "hello")

	if err := sync.Ensure(projectDir); err != nil {
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

	// State must be written with a switch path
	state, err := project.LoadState(projectDir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if state.SwitchPath == "" {
		t.Error("state missing SwitchPath after sync")
	}
}

// TestIntegration_SwitchPathStoredInState verifies that after sync the switch
// path is persisted in .oc/state.toml so subsequent calls stay on the same switch.
func TestIntegration_SwitchPathStoredInState(t *testing.T) {
	requireOpam(t)
	t.Log("Creating a switch to verify state persistence — may take several minutes on first run.")

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}
	projectDir := filepath.Join(dir, "hello")

	if err := sync.Ensure(projectDir); err != nil {
		t.Fatalf("first sync.Ensure: %v", err)
	}

	s1, _ := project.LoadState(projectDir)
	path1 := s1.SwitchPath

	// Second sync — should reuse stored path.
	if err := sync.Ensure(projectDir); err != nil {
		t.Fatalf("second sync.Ensure: %v", err)
	}

	s2, _ := project.LoadState(projectDir)
	path2 := s2.SwitchPath

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

	if err := sync.Ensure(projectDir); err != nil {
		t.Fatalf("sync.Ensure: %v", err)
	}

	state, err := project.LoadState(projectDir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if state.SwitchPath == "" {
		t.Fatal("state.SwitchPath empty after sync")
	}

	out, err := runCmd(projectDir, "opam", "exec", "--switch", state.SwitchPath, "--", "dune", "build")
	if err != nil {
		t.Fatalf("dune build failed: %v\noutput: %s", err, out)
	}

	exe := filepath.Join(projectDir, "_build", "default", "bin", "main.exe")
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("expected built executable at %s: %v", exe, err)
	}
}

// TestIntegration_AddDependency verifies oc add installs a package and the
// switch has it available after sync.
func TestIntegration_AddDependency(t *testing.T) {
	requireOpam(t)
	t.Log("This test installs an opam package — may take several minutes.")

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "hello", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}
	projectDir := filepath.Join(dir, "hello")

	// Add a small dep with no C dependencies for speed.
	if err := dune.AddDep(projectDir, "stringext", "*"); err != nil {
		t.Fatalf("dune.AddDep: %v", err)
	}

	if err := sync.Ensure(projectDir); err != nil {
		t.Fatalf("sync.Ensure: %v", err)
	}

	state, err := project.LoadState(projectDir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	// Verify stringext is installed in the switch.
	var buf bytes.Buffer
	if err := ocExec.Run("opam", []string{
		"list", "--installed", "--short", "--columns=name",
		"--switch", state.SwitchPath,
	}, ocExec.Options{Stdout: &buf}); err != nil {
		t.Fatalf("opam list: %v", err)
	}
	if !strings.Contains(buf.String(), "stringext") {
		t.Errorf("stringext not installed in switch; opam list output:\n%s", buf.String())
	}
}

func runCmd(dir, name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}

package sync_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
)

type mockRunner struct {
	switches      map[string]bool
	createCalled  []string
	installCalled []string
	lockCalled    []string
	createErr     error
	installErr    error
}

func (m *mockRunner) SwitchExists(path string) bool {
	return m.switches[path]
}

func (m *mockRunner) CreateSwitch(path, ocamlVersion string) error {
	m.createCalled = append(m.createCalled, path)
	m.switches[path] = true
	return m.createErr
}

func (m *mockRunner) InstallDeps(dir, switchPath string) error {
	m.installCalled = append(m.installCalled, switchPath)
	return m.installErr
}

func (m *mockRunner) LockDeps(dir string) error {
	m.lockCalled = append(m.lockCalled, dir)
	return nil
}

func TestEnsureWith_CreatesSwitch_WhenMissing(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	if len(runner.createCalled) != 1 {
		t.Errorf("expected CreateSwitch called once, got %d", len(runner.createCalled))
	}
}

func TestEnsureWith_SkipsCreate_WhenSwitchExists(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	// first call creates
	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
	// second call should reuse
	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatalf("second EnsureWith: %v", err)
	}

	if len(runner.createCalled) != 1 {
		t.Errorf("expected CreateSwitch called once total, got %d", len(runner.createCalled))
	}
}

func TestEnsureWith_CreatesSymlink(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	link := filepath.Join(dir, ".ocaml")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf(".ocaml symlink not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error(".ocaml should be a symlink")
	}
}

func TestEnsureWith_CallsInstallDeps(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	if len(runner.installCalled) != 1 {
		t.Errorf("expected InstallDeps called once, got %d", len(runner.installCalled))
	}
}

func TestEnsureWith_CallsLockDeps(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	if len(runner.lockCalled) != 1 {
		t.Errorf("expected LockDeps called once, got %d", len(runner.lockCalled))
	}
}

func TestEnsureWith_SavesState(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	s, err := project.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if s.SwitchPath == "" {
		t.Error("expected SwitchPath to be saved in state")
	}
	if s.OCamlVersion != "5.2.0" {
		t.Errorf("OCamlVersion in state: got %q, want %q", s.OCamlVersion, "5.2.0")
	}
}

func TestEnsureWith_ReusesSwitchPathAcrossCalls(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	// First call: switch created, state written with switch path
	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
	s1, _ := project.LoadState(dir)
	path1 := s1.SwitchPath

	// Second call: should reuse stored path, not recompute
	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
	s2, _ := project.LoadState(dir)
	path2 := s2.SwitchPath

	if path1 != path2 {
		t.Errorf("switch path changed between calls: %q → %q", path1, path2)
	}
	if len(runner.createCalled) != 1 {
		t.Errorf("CreateSwitch should be called exactly once, got %d", len(runner.createCalled))
	}
}

func TestEnsureWith_PropagatesCreateError(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{
		switches:  map[string]bool{},
		createErr: fmt.Errorf("opam failed"),
	}

	err := sync.EnsureWith(dir, "5.2.0", runner)
	if err == nil {
		t.Fatal("expected error when CreateSwitch fails")
	}
}

func TestEnsureWith_NewSwitchOnOCamlVersionChange(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	// First call with 5.2.0
	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
	if len(runner.createCalled) != 1 {
		t.Fatalf("expected 1 CreateSwitch call, got %d", len(runner.createCalled))
	}
	path1 := runner.createCalled[0]

	// Second call with a different OCaml version — stored path must be discarded
	if err := sync.EnsureWith(dir, "5.3.0", runner); err != nil {
		t.Fatal(err)
	}
	if len(runner.createCalled) != 2 {
		t.Fatalf("expected 2 CreateSwitch calls, got %d", len(runner.createCalled))
	}
	path2 := runner.createCalled[1]

	if path1 == path2 {
		t.Error("expected different switch paths for different OCaml versions")
	}
}

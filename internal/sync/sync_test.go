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
	switches       map[string]bool
	switchVersions map[string]string // path → ocaml version
	createCalled   []string
	installCalled  []string
	lockCalled     []string
	cloneCalled    [][2]string // [source, dest] pairs
	cachedSwitches []string   // returned by ListCachedSwitches
	createErr      error
	installErr     error
	cloneErr       error
}

func (m *mockRunner) SwitchExists(path string) bool {
	return m.switches[path]
}

func (m *mockRunner) CreateSwitch(path, ocamlVersion string) error {
	m.createCalled = append(m.createCalled, path)
	if m.createErr != nil {
		return m.createErr
	}
	m.switches[path] = true
	if m.switchVersions == nil {
		m.switchVersions = map[string]string{}
	}
	m.switchVersions[path] = ocamlVersion
	return nil
}

func (m *mockRunner) InstallDeps(dir, switchPath string) error {
	m.installCalled = append(m.installCalled, switchPath)
	return m.installErr
}

func (m *mockRunner) LockDeps(dir string) error {
	m.lockCalled = append(m.lockCalled, dir)
	return nil
}

func (m *mockRunner) CloneSwitch(source, dest string) error {
	m.cloneCalled = append(m.cloneCalled, [2]string{source, dest})
	if m.cloneErr != nil {
		return m.cloneErr
	}
	m.switches[dest] = true
	if m.switchVersions == nil {
		m.switchVersions = map[string]string{}
	}
	m.switchVersions[dest] = m.switchVersions[source]
	return nil
}

func (m *mockRunner) GetSwitchOCamlVersion(switchPath string) (string, error) {
	if v, ok := m.switchVersions[switchPath]; ok {
		return v, nil
	}
	return "", fmt.Errorf("switch not found: %s", switchPath)
}

func (m *mockRunner) ListCachedSwitches() ([]string, error) {
	return m.cachedSwitches, nil
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

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
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

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
	s1, _ := project.LoadState(dir)
	path1 := s1.SwitchPath

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

	if err := sync.EnsureWith(dir, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
	if len(runner.createCalled) != 1 {
		t.Fatalf("expected 1 CreateSwitch call, got %d", len(runner.createCalled))
	}
	path1 := runner.createCalled[0]

	runner.cachedSwitches = []string{path1}

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

func TestEnsureWith_ClonesSwitch_WhenCompatibleBaseExists(t *testing.T) {
	dirA := t.TempDir()
	runnerA := &mockRunner{switches: map[string]bool{}}
	if err := sync.EnsureWith(dirA, "5.2.0", runnerA); err != nil {
		t.Fatal(err)
	}
	existingSwitch := runnerA.createCalled[0]

	dirB := t.TempDir()
	runner := &mockRunner{
		switches:       map[string]bool{existingSwitch: true},
		switchVersions: map[string]string{existingSwitch: "5.2.0"},
		cachedSwitches: []string{existingSwitch},
	}

	if err := sync.EnsureWith(dirB, "5.2.0", runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	if len(runner.cloneCalled) != 1 {
		t.Errorf("expected CloneSwitch called once, got %d", len(runner.cloneCalled))
	}
	if len(runner.createCalled) != 0 {
		t.Errorf("expected CreateSwitch not called when base exists, got %d calls", len(runner.createCalled))
	}
	if runner.cloneCalled[0][0] != existingSwitch {
		t.Errorf("clone source: got %q, want %q", runner.cloneCalled[0][0], existingSwitch)
	}
}

func TestEnsureWith_FallsBackToCreate_WhenCloneFails(t *testing.T) {
	dirA := t.TempDir()
	runnerA := &mockRunner{switches: map[string]bool{}}
	if err := sync.EnsureWith(dirA, "5.2.0", runnerA); err != nil {
		t.Fatal(err)
	}
	existingSwitch := runnerA.createCalled[0]

	dirB := t.TempDir()
	runner := &mockRunner{
		switches:       map[string]bool{existingSwitch: true},
		switchVersions: map[string]string{existingSwitch: "5.2.0"},
		cachedSwitches: []string{existingSwitch},
		cloneErr:       fmt.Errorf("opam switch copy failed"),
	}

	if err := sync.EnsureWith(dirB, "5.2.0", runner); err != nil {
		t.Fatalf("EnsureWith should succeed via fallback: %v", err)
	}

	if len(runner.createCalled) != 1 {
		t.Errorf("expected CreateSwitch called once as fallback, got %d", len(runner.createCalled))
	}
}

func TestEnsureWith_SkipsClone_WhenNoCompatibleBase(t *testing.T) {
	dirA := t.TempDir()
	runnerA := &mockRunner{switches: map[string]bool{}}
	if err := sync.EnsureWith(dirA, "5.2.0", runnerA); err != nil {
		t.Fatal(err)
	}
	existingSwitch := runnerA.createCalled[0]

	dirB := t.TempDir()
	runner := &mockRunner{
		switches:       map[string]bool{existingSwitch: true},
		switchVersions: map[string]string{existingSwitch: "5.2.0"},
		cachedSwitches: []string{existingSwitch},
	}

	if err := sync.EnsureWith(dirB, "5.3.0", runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	if len(runner.cloneCalled) != 0 {
		t.Errorf("expected no CloneSwitch for incompatible version, got %d", len(runner.cloneCalled))
	}
	if len(runner.createCalled) != 1 {
		t.Errorf("expected CreateSwitch called once, got %d", len(runner.createCalled))
	}
}

func TestEnsureWith_PerProjectSwitchPaths_DifferByProject(t *testing.T) {
	runner := &mockRunner{switches: map[string]bool{}}

	dirA := t.TempDir()
	if err := sync.EnsureWith(dirA, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}
	runner.cachedSwitches = []string{runner.createCalled[0]}

	dirB := t.TempDir()
	if err := sync.EnsureWith(dirB, "5.2.0", runner); err != nil {
		t.Fatal(err)
	}

	sA, _ := project.LoadState(dirA)
	sB, _ := project.LoadState(dirB)
	if sA.SwitchPath == sB.SwitchPath {
		t.Error("different projects should get different switch paths")
	}
}

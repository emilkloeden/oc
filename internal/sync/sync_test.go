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
	switches        map[string]bool
	createCalled    []string
	installCalled   []string
	installedPkgs   []project.Package
	createErr       error
	installErr      error
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

func (m *mockRunner) ListInstalled(switchPath string) ([]project.Package, error) {
	return m.installedPkgs, nil
}

func cfg(ocamlVer string) *project.Config {
	return &project.Config{
		Project:         project.ProjectMeta{Name: "test_app", Version: "0.1.0"},
		OCaml:           project.OCamlMeta{Version: ocamlVer},
		Dependencies:    map[string]string{},
		DevDependencies: map[string]string{},
	}
}

func TestEnsureWith_CreatesSwitch_WhenMissing(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, cfg("5.2.0"), runner); err != nil {
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
	sync.EnsureWith(dir, cfg("5.2.0"), runner)
	// second call should reuse
	if err := sync.EnsureWith(dir, cfg("5.2.0"), runner); err != nil {
		t.Fatalf("second EnsureWith: %v", err)
	}

	if len(runner.createCalled) != 1 {
		t.Errorf("expected CreateSwitch called once total, got %d", len(runner.createCalled))
	}
}

func TestEnsureWith_CreatesSymlink(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, cfg("5.2.0"), runner); err != nil {
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

	if err := sync.EnsureWith(dir, cfg("5.2.0"), runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	if len(runner.installCalled) != 1 {
		t.Errorf("expected InstallDeps called once, got %d", len(runner.installCalled))
	}
}

func TestEnsureWith_UpdatesLockfile(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{
		switches: map[string]bool{},
		installedPkgs: []project.Package{
			{Name: "cohttp", Version: "5.0.0"},
			{Name: "lwt", Version: "5.7.0"},
		},
	}

	if err := sync.EnsureWith(dir, cfg("5.2.0"), runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	lock, err := project.LoadLock(dir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if len(lock.Packages) != 2 {
		t.Errorf("expected 2 packages in lockfile, got %d", len(lock.Packages))
	}
	if lock.OCaml.Version != "5.2.0" {
		t.Errorf("ocaml version in lock: got %q", lock.OCaml.Version)
	}
}

func TestEnsureWith_StoresSwitchPathInLock(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{switches: map[string]bool{}}

	if err := sync.EnsureWith(dir, cfg("5.2.0"), runner); err != nil {
		t.Fatalf("EnsureWith: %v", err)
	}

	lock, err := project.LoadLock(dir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if lock.SwitchPath == "" {
		t.Error("expected SwitchPath to be stored in lockfile")
	}
}

func TestEnsureWith_ReusesSwitchPathAcrossCalls(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{
		switches: map[string]bool{},
		installedPkgs: []project.Package{
			{Name: "cohttp", Version: "5.0.0"},
		},
	}

	// First call: switch created, lock written with packages + switch path
	sync.EnsureWith(dir, cfg("5.2.0"), runner)

	lock1, _ := project.LoadLock(dir)
	path1 := lock1.SwitchPath

	// Second call: packages now in lock → hash would differ without stored path
	sync.EnsureWith(dir, cfg("5.2.0"), runner)

	lock2, _ := project.LoadLock(dir)
	path2 := lock2.SwitchPath

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

	err := sync.EnsureWith(dir, cfg("5.2.0"), runner)
	if err == nil {
		t.Fatal("expected error when CreateSwitch fails")
	}
}

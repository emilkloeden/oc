package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emilkloeden/oc/internal/project"
)

func TestLoadConfig_Basic(t *testing.T) {
	dir := t.TempDir()
	content := `
[project]
name = "my_app"
version = "0.1.0"

[ocaml]
version = "5.2.0"

[dependencies]
cohttp = ">=5.0.0"

[dev-dependencies]
alcotest = "*"
`
	if err := os.WriteFile(filepath.Join(dir, "oc.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := project.LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Project.Name != "my_app" {
		t.Errorf("name: got %q want %q", cfg.Project.Name, "my_app")
	}
	if cfg.Project.Version != "0.1.0" {
		t.Errorf("version: got %q want %q", cfg.Project.Version, "0.1.0")
	}
	if cfg.OCaml.Version != "5.2.0" {
		t.Errorf("ocaml version: got %q want %q", cfg.OCaml.Version, "5.2.0")
	}
	if v, ok := cfg.Dependencies["cohttp"]; !ok || v != ">=5.0.0" {
		t.Errorf("dependencies: got %v", cfg.Dependencies)
	}
	if v, ok := cfg.DevDependencies["alcotest"]; !ok || v != "*" {
		t.Errorf("dev-dependencies: got %v", cfg.DevDependencies)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := project.LoadConfig(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing oc.toml")
	}
}

func TestLoadConfig_MissingName(t *testing.T) {
	dir := t.TempDir()
	content := `
[project]
version = "0.1.0"
[ocaml]
version = "5.2.0"
`
	if err := os.WriteFile(filepath.Join(dir, "oc.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := project.LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error when project name is missing")
	}
}

func TestSaveConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := &project.Config{
		Project: project.ProjectMeta{Name: "round_trip", Version: "1.0.0"},
		OCaml:   project.OCamlMeta{Version: "5.2.0"},
		Dependencies: map[string]string{
			"lwt": ">=5.7.0",
		},
		DevDependencies: map[string]string{},
	}

	if err := project.SaveConfig(dir, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got, err := project.LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig after save: %v", err)
	}
	if got.Project.Name != cfg.Project.Name {
		t.Errorf("name: got %q want %q", got.Project.Name, cfg.Project.Name)
	}
	if got.OCaml.Version != cfg.OCaml.Version {
		t.Errorf("ocaml version: got %q want %q", got.OCaml.Version, cfg.OCaml.Version)
	}
	if v, ok := got.Dependencies["lwt"]; !ok || v != ">=5.7.0" {
		t.Errorf("dependencies after round-trip: %v", got.Dependencies)
	}
}

func TestLoadLock_Basic(t *testing.T) {
	dir := t.TempDir()
	content := `
[ocaml]
version = "5.2.0"

[[package]]
name = "cohttp"
version = "5.0.0"

[[package]]
name = "lwt"
version = "5.7.0"
`
	if err := os.WriteFile(filepath.Join(dir, "oc.lock"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lock, err := project.LoadLock(dir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if lock.OCaml.Version != "5.2.0" {
		t.Errorf("ocaml version: got %q", lock.OCaml.Version)
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(lock.Packages))
	}
	if lock.Packages[0].Name != "cohttp" || lock.Packages[0].Version != "5.0.0" {
		t.Errorf("package[0]: %+v", lock.Packages[0])
	}
}

func TestLoadLock_Missing(t *testing.T) {
	lock, err := project.LoadLock(t.TempDir())
	if err != nil {
		t.Fatalf("LoadLock missing file should return empty lock, got error: %v", err)
	}
	if len(lock.Packages) != 0 {
		t.Errorf("expected empty packages, got %v", lock.Packages)
	}
}

func TestSaveLock_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	lock := &project.Lock{
		OCaml: project.OCamlMeta{Version: "5.2.0"},
		Packages: []project.Package{
			{Name: "cohttp", Version: "5.0.0"},
			{Name: "lwt", Version: "5.7.0"},
		},
	}
	if err := project.SaveLock(dir, lock); err != nil {
		t.Fatalf("SaveLock: %v", err)
	}
	got, err := project.LoadLock(dir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if len(got.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(got.Packages))
	}
	if got.Packages[1].Name != "lwt" {
		t.Errorf("package[1]: %+v", got.Packages[1])
	}
}

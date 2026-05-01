package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/cmd"
	"github.com/emilkloeden/oc/internal/project"
)

func TestProjectRoot_FindsDuneProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte("(lang dune 3.0)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	root, err := cmd.FindProjectRoot(subdir)
	if err != nil {
		t.Fatalf("cmd.FindProjectRoot: %v", err)
	}
	if root != dir {
		t.Errorf("got %q want %q", root, dir)
	}
}

func TestFormatEnvOutput_ShowsOCamlVersion(t *testing.T) {
	out := cmd.FormatEnvOutput("5.2.0", "")
	if !strings.Contains(out, "5.2.0") {
		t.Errorf("expected OCaml version in output, got:\n%s", out)
	}
}

func TestFormatEnvOutput_ShowsSwitchPath(t *testing.T) {
	out := cmd.FormatEnvOutput("5.2.0", "/cache/oc/switches/abc123/")
	if !strings.Contains(out, "/cache/oc/switches/abc123/") {
		t.Errorf("expected switch path in output, got:\n%s", out)
	}
}

func TestFormatEnvOutput_ShowsUninitialised_WhenNoSwitchPath(t *testing.T) {
	out := cmd.FormatEnvOutput("5.2.0", "")
	if !strings.Contains(out, "not yet initialised") {
		t.Errorf("expected uninitialised message in output, got:\n%s", out)
	}
}

func TestEnvState_LoadsFromStateFile(t *testing.T) {
	dir := t.TempDir()
	if err := project.SaveState(dir, project.State{
		SwitchPath:   "/cache/switch",
		OCamlVersion: "5.2.0",
	}); err != nil {
		t.Fatal(err)
	}
	s, err := project.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if s.SwitchPath != "/cache/switch" {
		t.Errorf("SwitchPath: got %q", s.SwitchPath)
	}
}

package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/cmd"
	"github.com/emilkloeden/oc/internal/project"
)

// loadEnvOutput scaffolds an oc project in a temp dir and captures the env
// command output by calling the formatting logic directly.
func loadEnvOutput(t *testing.T, lock *project.Lock) string {
	t.Helper()
	var buf bytes.Buffer
	cmd.PrintEnvInfo(&buf, lock)
	return buf.String()
}

func TestEnvOutput_ShowsOCamlVersion(t *testing.T) {
	lock := &project.Lock{
		OCaml:    project.OCamlMeta{Version: "5.2.0"},
		Packages: []project.Package{},
	}
	out := loadEnvOutput(t, lock)
	if !strings.Contains(out, "5.2.0") {
		t.Errorf("expected OCaml version in output, got:\n%s", out)
	}
}

func TestEnvOutput_ShowsPackages(t *testing.T) {
	lock := &project.Lock{
		OCaml: project.OCamlMeta{Version: "5.2.0"},
		Packages: []project.Package{
			{Name: "cohttp", Version: "5.0.0"},
			{Name: "lwt", Version: "5.7.0"},
		},
	}
	out := loadEnvOutput(t, lock)
	if !strings.Contains(out, "cohttp") {
		t.Errorf("expected cohttp in output, got:\n%s", out)
	}
	if !strings.Contains(out, "lwt") {
		t.Errorf("expected lwt in output, got:\n%s", out)
	}
}

func TestEnvOutput_EmptyPackages(t *testing.T) {
	lock := &project.Lock{
		OCaml:    project.OCamlMeta{Version: "5.2.0"},
		Packages: []project.Package{},
	}
	out := loadEnvOutput(t, lock)
	if !strings.Contains(out, "no packages") && !strings.Contains(out, "0 package") {
		t.Errorf("expected empty package message, got:\n%s", out)
	}
}

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

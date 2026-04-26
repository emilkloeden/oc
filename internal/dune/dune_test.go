package dune_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/internal/dune"
)

func TestScaffoldBin_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := dune.ScaffoldBin(dir, "my_app"); err != nil {
		t.Fatalf("ScaffoldBin: %v", err)
	}

	for _, f := range []string{"dune-project", "bin/dune", "bin/main.ml"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}
}

func TestScaffoldBin_DuneProject(t *testing.T) {
	dir := t.TempDir()
	if err := dune.ScaffoldBin(dir, "my_app"); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "dune-project"))
	if !strings.Contains(content, "(lang dune") {
		t.Error("dune-project missing lang stanza")
	}
	if !strings.Contains(content, `(name my_app)`) {
		t.Error("dune-project missing project name")
	}
}

func TestScaffoldBin_BinDune(t *testing.T) {
	dir := t.TempDir()
	if err := dune.ScaffoldBin(dir, "my_app"); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "bin/dune"))
	if !strings.Contains(content, "(executable") {
		t.Error("bin/dune missing executable stanza")
	}
	if !strings.Contains(content, "(name main)") {
		t.Error("bin/dune missing name")
	}
}

func TestScaffoldBin_MainML(t *testing.T) {
	dir := t.TempDir()
	if err := dune.ScaffoldBin(dir, "my_app"); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "bin/main.ml"))
	if !strings.Contains(content, `print_endline`) {
		t.Error("main.ml should have a hello world print")
	}
}

func TestScaffoldLib_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := dune.ScaffoldLib(dir, "my_lib"); err != nil {
		t.Fatalf("ScaffoldLib: %v", err)
	}

	for _, f := range []string{"dune-project", "lib/dune", "lib/my_lib.ml"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}
}

func TestScaffoldLib_LibDune(t *testing.T) {
	dir := t.TempDir()
	if err := dune.ScaffoldLib(dir, "my_lib"); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(dir, "lib/dune"))
	if !strings.Contains(content, "(library") {
		t.Error("lib/dune missing library stanza")
	}
	if !strings.Contains(content, "(name my_lib)") {
		t.Error("lib/dune missing name")
	}
}

func TestScaffoldBin_IdempotentDuneProject(t *testing.T) {
	dir := t.TempDir()
	if err := dune.ScaffoldBin(dir, "my_app"); err != nil {
		t.Fatal(err)
	}
	// calling again should not error
	if err := dune.ScaffoldBin(dir, "my_app"); err != nil {
		t.Fatalf("second ScaffoldBin should be idempotent: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile %s: %v", path, err)
	}
	return string(b)
}

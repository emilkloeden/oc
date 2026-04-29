package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/cmd"
)

func TestNew_InitializesGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_app", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	projectDir := filepath.Join(dir, "my_app")
	gitDir := filepath.Join(projectDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		t.Errorf("expected .git directory at %s: %v", gitDir, err)
	}
}

func TestNew_GitignoreExcludesOcamlDir(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_app", false); err != nil {
		t.Fatal(err)
	}

	gitignore := filepath.Join(dir, "my_app", ".gitignore")
	content, err := os.ReadFile(gitignore)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	for _, entry := range []string{".ocaml/", "_build/"} {
		found := false
		for _, line := range splitLines(string(content)) {
			if line == entry {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(".gitignore missing %q", entry)
		}
	}
}

func TestNew_CreatesExpectedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_app", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	projectDir := filepath.Join(dir, "my_app")
	for _, f := range []string{"dune-project", "bin/dune", "bin/main.ml", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(projectDir, f)); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}
}

func TestNew_DoesNotCreateOcToml(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_app", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	ocToml := filepath.Join(dir, "my_app", "oc.toml")
	if _, err := os.Stat(ocToml); err == nil {
		t.Error("oc.toml should not be created by oc new")
	}
}

func TestNew_DuneProjectHasGenerateOpamFiles(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_app", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "my_app", "dune-project"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "generate_opam_files") {
		t.Errorf("dune-project missing (generate_opam_files true):\n%s", content)
	}
}

func TestNew_DuneProjectHasPackageStanza(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_app", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "my_app", "dune-project"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "(package") {
		t.Errorf("dune-project missing (package ...) stanza:\n%s", content)
	}
	if !strings.Contains(string(content), "my_app") {
		t.Errorf("dune-project package stanza missing name my_app:\n%s", content)
	}
}

func TestNew_CreatesInitialLockWithOcamlVersion(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_app", false); err != nil {
		t.Fatalf("RunNew: %v", err)
	}

	lockPath := filepath.Join(dir, "my_app", "oc.lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("expected oc.lock to exist: %v", err)
	}
	content, _ := os.ReadFile(lockPath)
	if !strings.Contains(string(content), "5.2.0") {
		t.Errorf("expected oc.lock to contain default OCaml version 5.2.0:\n%s", content)
	}
}

func TestNew_LibFlag(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunNew(dir, "my_lib", true); err != nil {
		t.Fatalf("RunNew --lib: %v", err)
	}

	projectDir := filepath.Join(dir, "my_lib")
	for _, f := range []string{"lib/dune", "lib/my_lib.ml"} {
		if _, err := os.Stat(filepath.Join(projectDir, f)); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}
}

func TestNew_FailsIfDirExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "my_app"), 0755); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunNew(dir, "my_app", false)
	if err == nil {
		t.Fatal("expected error when project directory already exists")
	}
}

func TestNew_InvalidNameReturnsError(t *testing.T) {
	dir := t.TempDir()
	cases := []string{
		"my project",  // space
		"123abc",      // starts with digit
		"my-project",  // hyphen
		"my.project",  // dot
		"my(project",  // parens
		"",            // empty
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			err := cmd.RunNew(dir, name, false)
			if err == nil {
				t.Errorf("expected error for name %q, got nil", name)
			}
		})
	}
}

func TestNew_ValidNameSucceeds(t *testing.T) {
	cases := []string{
		"myapp",
		"my_app",
		"MyApp",
		"app123",
		"a",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			subDir := t.TempDir()
			if err := cmd.RunNew(subDir, name, false); err != nil {
				t.Errorf("unexpected error for name %q: %v", name, err)
			}
		})
	}
}

func TestNew_InvalidNameDoesNotCreateDir(t *testing.T) {
	dir := t.TempDir()
	name := "invalid name with spaces"
	_ = cmd.RunNew(dir, name, false)
	projectDir := filepath.Join(dir, name)
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Errorf("expected no directory at %s after failed RunNew", projectDir)
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

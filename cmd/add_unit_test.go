package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/cmd"
	"github.com/emilkloeden/oc/internal/project"
)

const addTestDuneProject = `(lang dune 3.0)
(generate_opam_files true)

(package
 (name my_app)
 (synopsis "test")
 (depends
  (ocaml (>= "5.2.0"))
  dune))
`

const addTestOpamFile = `opam-version: "2.0"
name: "my_app"
depends: [
  "ocaml" {>= "5.2.0"}
  "dune" {>= "3.0"}
]
`

func noopSync(_ string) error { return nil }

func TestRunAdd_DuneManagedProject_AddsToduneProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte(addTestDuneProject), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []project.Dep{{Name: "yojson", Constraint: "*"}}
	if err := cmd.RunAdd(dir, deps, noopSync); err != nil {
		t.Fatalf("RunAdd: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "dune-project"))
	if !strings.Contains(string(content), "yojson") {
		t.Errorf("expected yojson in dune-project:\n%s", content)
	}
}

func TestRunAdd_DuneManagedProject_AddsWithConstraint(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte(addTestDuneProject), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []project.Dep{{Name: "cohttp", Constraint: ">=5.0.0"}}
	if err := cmd.RunAdd(dir, deps, noopSync); err != nil {
		t.Fatalf("RunAdd: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "dune-project"))
	if !strings.Contains(string(content), "cohttp") || !strings.Contains(string(content), "5.0.0") {
		t.Errorf("expected cohttp with constraint in dune-project:\n%s", content)
	}
}

func TestRunAdd_HandWrittenOpam_AddsToOpamFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "my_app.opam"), []byte(addTestOpamFile), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []project.Dep{{Name: "yojson", Constraint: "*"}}
	if err := cmd.RunAdd(dir, deps, noopSync); err != nil {
		t.Fatalf("RunAdd: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "my_app.opam"))
	if !strings.Contains(string(content), `"yojson"`) {
		t.Errorf("expected yojson in opam file:\n%s", content)
	}
}

func TestRunAdd_DuneManagedProject_IdempotentAdd(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte(addTestDuneProject), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []project.Dep{{Name: "yojson", Constraint: "*"}}
	if err := cmd.RunAdd(dir, deps, noopSync); err != nil {
		t.Fatalf("first RunAdd: %v", err)
	}
	if err := cmd.RunAdd(dir, deps, noopSync); err != nil {
		t.Fatalf("second RunAdd: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "dune-project"))
	count := strings.Count(string(content), "yojson")
	if count != 1 {
		t.Errorf("expected yojson exactly once after idempotent add, got %d:\n%s", count, content)
	}
}

func TestRunAdd_SyncFuncCalledWithProjectDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte(addTestDuneProject), 0644); err != nil {
		t.Fatal(err)
	}

	var syncCalledWith string
	captureSync := func(d string) error {
		syncCalledWith = d
		return nil
	}

	deps := []project.Dep{{Name: "yojson", Constraint: "*"}}
	if err := cmd.RunAdd(dir, deps, captureSync); err != nil {
		t.Fatalf("RunAdd: %v", err)
	}

	if syncCalledWith != dir {
		t.Errorf("sync called with %q, want %q", syncCalledWith, dir)
	}
}

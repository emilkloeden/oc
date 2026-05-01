package dune_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/internal/dune"
)

const sampleDuneProject = `(lang dune 3.0)
(generate_opam_files true)

(package
 (name my_app)
 (synopsis "A test application")
 (depends
  (ocaml (>= "5.2.0"))
  dune))
`

func writeDuneProject(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readDuneProject(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "dune-project"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAddDep_AddsPackageWithNoConstraint(t *testing.T) {
	dir := t.TempDir()
	writeDuneProject(t, dir, sampleDuneProject)

	if err := dune.AddDep(dir, "yojson", "*"); err != nil {
		t.Fatalf("AddDep: %v", err)
	}

	content := readDuneProject(t, dir)
	if !strings.Contains(content, "yojson") {
		t.Errorf("expected yojson in dune-project:\n%s", content)
	}
}

func TestAddDep_AddsPackageWithConstraint(t *testing.T) {
	dir := t.TempDir()
	writeDuneProject(t, dir, sampleDuneProject)

	if err := dune.AddDep(dir, "cohttp", ">=5.0.0"); err != nil {
		t.Fatalf("AddDep: %v", err)
	}

	content := readDuneProject(t, dir)
	if !strings.Contains(content, "cohttp") {
		t.Errorf("expected cohttp in dune-project:\n%s", content)
	}
	if !strings.Contains(content, "5.0.0") {
		t.Errorf("expected version 5.0.0 in dune-project:\n%s", content)
	}
}

func TestAddDep_IdempotentWhenPackageAlreadyPresent(t *testing.T) {
	dir := t.TempDir()
	writeDuneProject(t, dir, sampleDuneProject)

	if err := dune.AddDep(dir, "yojson", "*"); err != nil {
		t.Fatalf("first AddDep: %v", err)
	}
	if err := dune.AddDep(dir, "yojson", "*"); err != nil {
		t.Fatalf("second AddDep: %v", err)
	}

	content := readDuneProject(t, dir)
	// Should appear exactly once
	count := strings.Count(content, "yojson")
	if count != 1 {
		t.Errorf("expected yojson to appear exactly once, got %d:\n%s", count, content)
	}
}

func TestAddDep_ErrorWhenNoDuneProject(t *testing.T) {
	dir := t.TempDir()
	err := dune.AddDep(dir, "yojson", "*")
	if err == nil {
		t.Fatal("expected error when dune-project is missing")
	}
}

func TestRemoveDep_RemovesPackage(t *testing.T) {
	dir := t.TempDir()
	// Start with yojson already added
	content := `(lang dune 3.0)
(generate_opam_files true)

(package
 (name my_app)
 (synopsis "A test application")
 (depends
  (ocaml (>= "5.2.0"))
  dune
  yojson))
`
	writeDuneProject(t, dir, content)

	if err := dune.RemoveDep(dir, "yojson"); err != nil {
		t.Fatalf("RemoveDep: %v", err)
	}

	result := readDuneProject(t, dir)
	if strings.Contains(result, "yojson") {
		t.Errorf("expected yojson to be removed:\n%s", result)
	}
}

func TestRemoveDep_RemovesConstrainedPackage(t *testing.T) {
	dir := t.TempDir()
	content := `(lang dune 3.0)
(generate_opam_files true)

(package
 (name my_app)
 (synopsis "A test application")
 (depends
  (ocaml (>= "5.2.0"))
  dune
  (cohttp (>= "5.0.0"))))
`
	writeDuneProject(t, dir, content)

	if err := dune.RemoveDep(dir, "cohttp"); err != nil {
		t.Fatalf("RemoveDep: %v", err)
	}

	result := readDuneProject(t, dir)
	if strings.Contains(result, "cohttp") {
		t.Errorf("expected cohttp to be removed:\n%s", result)
	}
}

func TestRemoveDep_ErrorWhenPackageNotPresent(t *testing.T) {
	dir := t.TempDir()
	writeDuneProject(t, dir, sampleDuneProject)

	err := dune.RemoveDep(dir, "notexist")
	if err == nil {
		t.Fatal("expected error when package not in depends")
	}
}

func TestGetPackageName_ReturnsName(t *testing.T) {
	dir := t.TempDir()
	writeDuneProject(t, dir, sampleDuneProject)

	name, err := dune.GetPackageName(dir)
	if err != nil {
		t.Fatalf("GetPackageName: %v", err)
	}
	if name != "my_app" {
		t.Errorf("got %q, want %q", name, "my_app")
	}
}

func TestGetPackageName_NameAfterSynopsisWithParens(t *testing.T) {
	// synopsis contains "(name foo)" — an unbounded Index would return "foo"
	// instead of the real package name when name stanza follows synopsis.
	dir := t.TempDir()
	content := `(lang dune 3.0)
(generate_opam_files true)

(package
 (synopsis "Fixes (name resolution) issues")
 (name my_app)
 (depends dune))
`
	writeDuneProject(t, dir, content)
	name, err := dune.GetPackageName(dir)
	if err != nil {
		t.Fatalf("GetPackageName: %v", err)
	}
	if name != "my_app" {
		t.Errorf("got %q, want %q", name, "my_app")
	}
}

func TestGetPackageName_NameNotFoundInBlock(t *testing.T) {
	// A package stanza without (name ...) — should return an error,
	// not silently return a name from later in the file.
	dir := t.TempDir()
	content := `(lang dune 3.0)
(generate_opam_files true)

(package
 (synopsis "no name here")
 (depends dune))
`
	writeDuneProject(t, dir, content)
	_, err := dune.GetPackageName(dir)
	if err == nil {
		t.Fatal("expected error when (name ...) is absent from the package block")
	}
}

func TestHasGenerateOpamFiles_TrueWhenPresent(t *testing.T) {
	dir := t.TempDir()
	writeDuneProject(t, dir, sampleDuneProject)

	if !dune.HasGenerateOpamFiles(dir) {
		t.Error("expected HasGenerateOpamFiles to return true")
	}
}

func TestHasGenerateOpamFiles_FalseWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	content := "(lang dune 3.0)\n(name my_app)\n"
	writeDuneProject(t, dir, content)

	if dune.HasGenerateOpamFiles(dir) {
		t.Error("expected HasGenerateOpamFiles to return false")
	}
}

func writeDuneProjectBytes(t *testing.T, dir string, content []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "dune-project"), content, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestAddDep_IdempotentWithCRLFLineEndings(t *testing.T) {
	// CRLF-encoded dune-project files (Windows or git autocrlf) must not cause
	// hasDep to miss existing entries, which would produce duplicate deps.
	dir := t.TempDir()
	lf := "(lang dune 3.0)\r\n(generate_opam_files true)\r\n\r\n(package\r\n (name my_app)\r\n (depends\r\n  dune\r\n  yojson))\r\n"
	writeDuneProjectBytes(t, dir, []byte(lf))

	if err := dune.AddDep(dir, "yojson", "*"); err != nil {
		t.Fatalf("AddDep: %v", err)
	}

	result := readDuneProject(t, dir)
	count := strings.Count(result, "yojson")
	if count != 1 {
		t.Errorf("expected yojson exactly once after idempotent add on CRLF file, got %d:\n%s", count, result)
	}
}

func TestAddDep_IdempotentCRLFConstrainedDep(t *testing.T) {
	// Same as above but for a constrained dep entry like (cohttp (>= "5.0.0"))
	// where the atom inside the list must be read correctly past CRLF whitespace.
	dir := t.TempDir()
	lf := "(lang dune 3.0)\r\n(generate_opam_files true)\r\n\r\n(package\r\n (name my_app)\r\n (depends\r\n  dune\r\n  (cohttp (>= \"5.0.0\"))))\r\n"
	writeDuneProjectBytes(t, dir, []byte(lf))

	if err := dune.AddDep(dir, "cohttp", ">=5.0.0"); err != nil {
		t.Fatalf("AddDep: %v", err)
	}

	result := readDuneProject(t, dir)
	count := strings.Count(result, "cohttp")
	if count != 1 {
		t.Errorf("expected cohttp exactly once, got %d:\n%s", count, result)
	}
}

func TestAddDep_IdempotentCRLFMultiLineConstrainedDep(t *testing.T) {
	// A constrained dep whose name and constraint are on separate lines (CRLF):
	//   (cohttp\r\n   (>= "5.0.0"))
	// containsDepName reads the first atom inside the list; if '\r' is not a
	// terminator, it reads "cohttp\r" instead of "cohttp" and misses the match.
	dir := t.TempDir()
	// dune-project with CRLF where cohttp's name and constraint span two lines
	raw := "(lang dune 3.0)\r\n(generate_opam_files true)\r\n\r\n(package\r\n (name my_app)\r\n (depends\r\n  dune\r\n  (cohttp\r\n   (>= \"5.0.0\"))))\r\n"
	writeDuneProjectBytes(t, dir, []byte(raw))

	if err := dune.AddDep(dir, "cohttp", ">=5.0.0"); err != nil {
		t.Fatalf("AddDep: %v", err)
	}

	result := readDuneProject(t, dir)
	count := strings.Count(result, "cohttp")
	if count != 1 {
		t.Errorf("expected cohttp exactly once after idempotent add, got %d:\n%q", count, result)
	}
}

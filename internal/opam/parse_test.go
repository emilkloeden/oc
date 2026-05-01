package opam_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/internal/opam"
)

const sampleOpam = `opam-version: "2.0"
name: "my_app"
version: "0.1.0"
synopsis: "A test app"
depends: [
  "ocaml" {>= "4.14"}
  "dune" {>= "3.0"}
]
`

const sampleOpamWithDep = `opam-version: "2.0"
name: "my_app"
version: "0.1.0"
synopsis: "A test app"
depends: [
  "ocaml" {>= "4.14"}
  "dune" {>= "3.0"}
  "yojson"
]
`

func writeOpam(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name+".opam")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindOpamFile_ReturnsPath(t *testing.T) {
	dir := t.TempDir()
	writeOpam(t, dir, "my_app", sampleOpam)

	path, err := opam.FindOpamFile(dir)
	if err != nil {
		t.Fatalf("FindOpamFile: %v", err)
	}
	if !strings.HasSuffix(path, "my_app.opam") {
		t.Errorf("got %q, want path ending in my_app.opam", path)
	}
}

func TestFindOpamFile_ErrorWhenNone(t *testing.T) {
	dir := t.TempDir()
	_, err := opam.FindOpamFile(dir)
	if err == nil {
		t.Fatal("expected error when no .opam file present")
	}
}

func TestAddDepToOpam_AddsWithNoConstraint(t *testing.T) {
	dir := t.TempDir()
	path := writeOpam(t, dir, "my_app", sampleOpam)

	if err := opam.AddDepToOpam(path, "yojson", "*"); err != nil {
		t.Fatalf("AddDepToOpam: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), `"yojson"`) {
		t.Errorf("expected yojson in opam file:\n%s", content)
	}
}

func TestAddDepToOpam_AddsWithConstraint(t *testing.T) {
	dir := t.TempDir()
	path := writeOpam(t, dir, "my_app", sampleOpam)

	if err := opam.AddDepToOpam(path, "cohttp", ">=5.0.0"); err != nil {
		t.Fatalf("AddDepToOpam: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), `"cohttp"`) {
		t.Errorf("expected cohttp in opam file:\n%s", content)
	}
	if !strings.Contains(string(content), "5.0.0") {
		t.Errorf("expected version 5.0.0 in opam file:\n%s", content)
	}
}

func TestAddDepToOpam_IdempotentWhenAlreadyPresent(t *testing.T) {
	dir := t.TempDir()
	path := writeOpam(t, dir, "my_app", sampleOpam)

	if err := opam.AddDepToOpam(path, "yojson", "*"); err != nil {
		t.Fatalf("first AddDepToOpam: %v", err)
	}
	if err := opam.AddDepToOpam(path, "yojson", "*"); err != nil {
		t.Fatalf("second AddDepToOpam: %v", err)
	}

	content, _ := os.ReadFile(path)
	count := strings.Count(string(content), `"yojson"`)
	if count != 1 {
		t.Errorf("expected yojson exactly once, got %d:\n%s", count, content)
	}
}

func TestRemoveDepFromOpam_RemovesPackage(t *testing.T) {
	dir := t.TempDir()
	path := writeOpam(t, dir, "my_app", sampleOpamWithDep)

	if err := opam.RemoveDepFromOpam(path, "yojson"); err != nil {
		t.Fatalf("RemoveDepFromOpam: %v", err)
	}

	content, _ := os.ReadFile(path)
	if strings.Contains(string(content), `"yojson"`) {
		t.Errorf("expected yojson to be removed:\n%s", content)
	}
}

func TestRemoveDepFromOpam_ErrorWhenNotPresent(t *testing.T) {
	dir := t.TempDir()
	path := writeOpam(t, dir, "my_app", sampleOpam)

	err := opam.RemoveDepFromOpam(path, "notexist")
	if err == nil {
		t.Fatal("expected error when package not in depends")
	}
}

// --- ReadOCamlVersion tests ---

func TestReadOCamlVersion_GeConstraint(t *testing.T) {
	dir := t.TempDir()
	writeOpam(t, dir, "my_app", sampleOpam) // contains "ocaml" {>= "4.14"}
	v, err := opam.ReadOCamlVersion(dir)
	if err != nil {
		t.Fatalf("ReadOCamlVersion: %v", err)
	}
	if v != "4.14" {
		t.Errorf("got %q, want %q", v, "4.14")
	}
}

func TestReadOCamlVersion_EqConstraint(t *testing.T) {
	dir := t.TempDir()
	content := `opam-version: "2.0"
depends: [
  "ocaml" {= "5.1.0"}
  "dune" {>= "3.0"}
]
`
	writeOpam(t, dir, "my_app", content)
	v, err := opam.ReadOCamlVersion(dir)
	if err != nil {
		t.Fatalf("ReadOCamlVersion: %v", err)
	}
	if v != "5.1.0" {
		t.Errorf("got %q, want %q", v, "5.1.0")
	}
}

func TestReadOCamlVersion_MissingConstraint(t *testing.T) {
	dir := t.TempDir()
	content := `opam-version: "2.0"
depends: [
  "dune" {>= "3.0"}
]
`
	writeOpam(t, dir, "my_app", content)
	_, err := opam.ReadOCamlVersion(dir)
	if err == nil {
		t.Fatal("expected error when no ocaml constraint found")
	}
}

func TestReadOCamlVersion_MissingOpamFile(t *testing.T) {
	_, err := opam.ReadOCamlVersion(t.TempDir())
	if err == nil {
		t.Fatal("expected error when no .opam file present")
	}
}

func TestReadOCamlVersion_ReadsFromDependsBlockOnly(t *testing.T) {
	// Verify the function scans only the depends block, not the full file.
	// A file with depopts before depends: the depopts block ends with ']' before
	// depends: opens. If start were incorrectly 0 the scanner would include content
	// from before the depends block.
	dir := t.TempDir()
	content := `opam-version: "2.0"
depopts: [
  "threads" {>= "0.1"}
]
depends: [
  "ocaml" {>= "5.2.0"}
  "dune" {>= "3.0"}
]
`
	writeOpam(t, dir, "my_app", content)
	v, err := opam.ReadOCamlVersion(dir)
	if err != nil {
		t.Fatalf("ReadOCamlVersion: %v", err)
	}
	if v != "5.2.0" {
		t.Errorf("got %q, want %q", v, "5.2.0")
	}
}

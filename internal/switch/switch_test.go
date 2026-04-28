package swmgr_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	sw "github.com/emilkloeden/oc/internal/switch"
	"github.com/emilkloeden/oc/internal/project"
)

func lock(ocamlVer string, pkgs ...project.Package) *project.Lock {
	return &project.Lock{
		OCaml:    project.OCamlMeta{Version: ocamlVer},
		Packages: pkgs,
	}
}

func TestHash_DeterministicSameInputs(t *testing.T) {
	l := lock("5.2.0",
		project.Package{Name: "cohttp", Version: "5.0.0"},
		project.Package{Name: "lwt", Version: "5.7.0"},
	)
	h1 := sw.Hash(l)
	h2 := sw.Hash(l)
	if h1 != h2 {
		t.Errorf("hash not deterministic: %q vs %q", h1, h2)
	}
}

func TestHash_DifferentOCamlVersion(t *testing.T) {
	l1 := lock("5.2.0", project.Package{Name: "cohttp", Version: "5.0.0"})
	l2 := lock("5.1.0", project.Package{Name: "cohttp", Version: "5.0.0"})
	if sw.Hash(l1) == sw.Hash(l2) {
		t.Error("different ocaml versions should produce different hashes")
	}
}

func TestHash_DifferentPackages(t *testing.T) {
	l1 := lock("5.2.0", project.Package{Name: "cohttp", Version: "5.0.0"})
	l2 := lock("5.2.0", project.Package{Name: "cohttp", Version: "5.1.0"})
	if sw.Hash(l1) == sw.Hash(l2) {
		t.Error("different package versions should produce different hashes")
	}
}

func TestHash_OrderIndependent(t *testing.T) {
	l1 := lock("5.2.0",
		project.Package{Name: "cohttp", Version: "5.0.0"},
		project.Package{Name: "lwt", Version: "5.7.0"},
	)
	l2 := lock("5.2.0",
		project.Package{Name: "lwt", Version: "5.7.0"},
		project.Package{Name: "cohttp", Version: "5.0.0"},
	)
	if sw.Hash(l1) != sw.Hash(l2) {
		t.Error("package order should not affect hash")
	}
}

func TestCachePath_ContainsHash(t *testing.T) {
	l := lock("5.2.0", project.Package{Name: "cohttp", Version: "5.0.0"})
	h := sw.Hash(l)
	path := sw.CachePath(l)
	if filepath.Base(path) != h {
		t.Errorf("CachePath base should be hash %q, got %q", h, filepath.Base(path))
	}
}

func TestEnsureSymlink_CreatesLink(t *testing.T) {
	projectDir := t.TempDir()
	target := t.TempDir() // simulates a switch directory that already exists

	if err := sw.EnsureSymlink(projectDir, target); err != nil {
		t.Fatalf("EnsureSymlink: %v", err)
	}

	link := filepath.Join(projectDir, ".ocaml")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error(".ocaml should be a symlink")
	}
	resolved, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != target {
		t.Errorf("symlink target: got %q want %q", resolved, target)
	}
}

func TestEnsureSymlink_UpdatesStaleLink(t *testing.T) {
	projectDir := t.TempDir()
	old := t.TempDir()
	newTarget := t.TempDir()

	// create initial symlink
	if err := os.Symlink(old, filepath.Join(projectDir, ".ocaml")); err != nil {
		t.Fatal(err)
	}

	if err := sw.EnsureSymlink(projectDir, newTarget); err != nil {
		t.Fatalf("EnsureSymlink update: %v", err)
	}

	resolved, _ := os.Readlink(filepath.Join(projectDir, ".ocaml"))
	if resolved != newTarget {
		t.Errorf("symlink not updated: got %q want %q", resolved, newTarget)
	}
}

func TestEnsureSymlink_RegularFileReturnsError(t *testing.T) {
	projectDir := t.TempDir()
	link := filepath.Join(projectDir, ".ocaml")
	// Create a regular file (not a symlink) at .ocaml
	if err := os.WriteFile(link, []byte("not a symlink"), 0644); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	err := sw.EnsureSymlink(projectDir, target)
	if err == nil {
		t.Fatal("expected error when .ocaml is a regular file, got nil")
	}
	if !strings.Contains(err.Error(), "remove it manually") {
		t.Errorf("error should mention 'remove it manually' to guide the user; got: %v", err)
	}
}

func TestEnsureSymlink_DirectoryReturnsError(t *testing.T) {
	projectDir := t.TempDir()
	link := filepath.Join(projectDir, ".ocaml")
	// Create a directory at .ocaml
	if err := os.MkdirAll(link, 0755); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	err := sw.EnsureSymlink(projectDir, target)
	if err == nil {
		t.Fatal("expected error when .ocaml is a directory, got nil")
	}
	if !strings.Contains(err.Error(), "remove it manually") {
		t.Errorf("error should mention 'remove it manually' to guide the user; got: %v", err)
	}
}

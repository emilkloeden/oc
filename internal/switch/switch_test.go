package swmgr_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	sw "github.com/emilkloeden/oc/internal/switch"
)

func TestCachePathForVersion_Deterministic(t *testing.T) {
	p1, err := sw.CachePathForVersion("5.2.0")
	if err != nil {
		t.Fatalf("CachePathForVersion: %v", err)
	}
	p2, err := sw.CachePathForVersion("5.2.0")
	if err != nil {
		t.Fatalf("CachePathForVersion: %v", err)
	}
	if p1 != p2 {
		t.Errorf("CachePathForVersion not deterministic: %q vs %q", p1, p2)
	}
}

func TestCachePathForVersion_DiffersForDifferentVersions(t *testing.T) {
	p1, err := sw.CachePathForVersion("5.2.0")
	if err != nil {
		t.Fatalf("CachePathForVersion: %v", err)
	}
	p2, err := sw.CachePathForVersion("5.3.0")
	if err != nil {
		t.Fatalf("CachePathForVersion: %v", err)
	}
	if p1 == p2 {
		t.Error("different OCaml versions should produce different cache paths")
	}
}

func TestCachePathForVersion_ContainsExpectedSegments(t *testing.T) {
	path, err := sw.CachePathForVersion("5.2.0")
	if err != nil {
		t.Fatalf("CachePathForVersion: %v", err)
	}
	if !strings.Contains(path, filepath.Join(".cache", "oc", "switches")) {
		t.Errorf("unexpected path structure: %q", path)
	}
	base := filepath.Base(path)
	if len(base) != 16 {
		t.Errorf("expected 16-char hash in path base, got %q (len %d)", base, len(base))
	}
}

func TestEnsureSymlink_CreatesLink(t *testing.T) {
	projectDir := t.TempDir()
	target := t.TempDir()

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
	if err := os.WriteFile(link, []byte("not a symlink"), 0644); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	err := sw.EnsureSymlink(projectDir, target)
	if err == nil {
		t.Fatal("expected error when .ocaml is a regular file, got nil")
	}
	if !strings.Contains(err.Error(), "remove it manually") {
		t.Errorf("error should mention 'remove it manually'; got: %v", err)
	}
}

func TestCachePathForProject_Deterministic(t *testing.T) {
	p1, err := sw.CachePathForProject("/home/user/myproject", "5.2.0")
	if err != nil {
		t.Fatalf("CachePathForProject: %v", err)
	}
	p2, err := sw.CachePathForProject("/home/user/myproject", "5.2.0")
	if err != nil {
		t.Fatalf("CachePathForProject: %v", err)
	}
	if p1 != p2 {
		t.Errorf("CachePathForProject not deterministic: %q vs %q", p1, p2)
	}
}

func TestCachePathForProject_DiffersForDifferentVersions(t *testing.T) {
	p1, err := sw.CachePathForProject("/home/user/myproject", "5.2.0")
	if err != nil {
		t.Fatalf("CachePathForProject: %v", err)
	}
	p2, err := sw.CachePathForProject("/home/user/myproject", "5.3.0")
	if err != nil {
		t.Fatalf("CachePathForProject: %v", err)
	}
	if p1 == p2 {
		t.Error("different OCaml versions should produce different cache paths")
	}
}

func TestCachePathForProject_DiffersForDifferentProjectDirs(t *testing.T) {
	p1, err := sw.CachePathForProject("/home/user/projectA", "5.2.0")
	if err != nil {
		t.Fatalf("CachePathForProject: %v", err)
	}
	p2, err := sw.CachePathForProject("/home/user/projectB", "5.2.0")
	if err != nil {
		t.Fatalf("CachePathForProject: %v", err)
	}
	if p1 == p2 {
		t.Error("different project dirs should produce different cache paths")
	}
}

func TestCachePathForProject_ContainsExpectedSegments(t *testing.T) {
	path, err := sw.CachePathForProject("/home/user/myproject", "5.2.0")
	if err != nil {
		t.Fatalf("CachePathForProject: %v", err)
	}
	if !strings.Contains(path, filepath.Join(".cache", "oc", "switches")) {
		t.Errorf("unexpected path structure: %q", path)
	}
	base := filepath.Base(path)
	if len(base) != 16 {
		t.Errorf("expected 16-char hash in path base, got %q (len %d)", base, len(base))
	}
}

func TestListCachedSwitches_EmptyWhenNone(t *testing.T) {
	switches, err := sw.ListCachedSwitches(t.TempDir())
	if err != nil {
		t.Fatalf("ListCachedSwitches: %v", err)
	}
	if len(switches) != 0 {
		t.Errorf("expected empty list, got %v", switches)
	}
}

func TestListCachedSwitches_ReturnsDirectories(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "abc123"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(base, "def456"), 0755); err != nil {
		t.Fatal(err)
	}

	switches, err := sw.ListCachedSwitches(base)
	if err != nil {
		t.Fatalf("ListCachedSwitches: %v", err)
	}
	if len(switches) != 2 {
		t.Errorf("expected 2 switches, got %d: %v", len(switches), switches)
	}
}

func TestEnsureSymlink_DirectoryReturnsError(t *testing.T) {
	projectDir := t.TempDir()
	link := filepath.Join(projectDir, ".ocaml")
	if err := os.MkdirAll(link, 0755); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	err := sw.EnsureSymlink(projectDir, target)
	if err == nil {
		t.Fatal("expected error when .ocaml is a directory, got nil")
	}
	if !strings.Contains(err.Error(), "remove it manually") {
		t.Errorf("error should mention 'remove it manually'; got: %v", err)
	}
}

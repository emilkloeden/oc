package project_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/emilkloeden/oc/internal/project"
)

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

func TestLoadLock_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "oc.lock"), []byte("NOT VALID TOML ][[["), 0644); err != nil {
		t.Fatal(err)
	}
	lock, err := project.LoadLock(dir)
	if err == nil {
		t.Fatal("expected error for corrupted oc.lock, got nil")
	}
	if lock != nil {
		t.Errorf("expected nil lock on error, got %+v", lock)
	}
}

func TestSaveLock_NoTempFilesLeftOnSuccess(t *testing.T) {
	dir := t.TempDir()
	lock := &project.Lock{OCaml: project.OCamlMeta{Version: "5.2.0"}}
	if err := project.SaveLock(dir, lock); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".oc.lock.") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind after SaveLock: %s", e.Name())
		}
	}
}

func TestSaveLock_ConcurrentWritesProduceValidFile(t *testing.T) {
	dir := t.TempDir()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			lock := &project.Lock{
				OCaml:      project.OCamlMeta{Version: "5.2.0"},
				SwitchPath: fmt.Sprintf("/path/to/switch/%d", i),
				Packages:   []project.Package{{Name: "pkg", Version: fmt.Sprintf("1.0.%d", i)}},
			}
			_ = project.SaveLock(dir, lock)
		}()
	}
	wg.Wait()

	_, err := project.LoadLock(dir)
	if err != nil {
		t.Errorf("after concurrent SaveLock, oc.lock is not parseable: %v", err)
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

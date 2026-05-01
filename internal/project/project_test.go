package project_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	gosync "sync"
	"testing"

	"github.com/emilkloeden/oc/internal/project"
)

func TestLoadState_Missing(t *testing.T) {
	s, err := project.LoadState(t.TempDir())
	if err != nil {
		t.Fatalf("LoadState missing file should return empty State, got error: %v", err)
	}
	if s.SwitchPath != "" {
		t.Errorf("expected empty SwitchPath, got %q", s.SwitchPath)
	}
	if s.OCamlVersion != "" {
		t.Errorf("expected empty OCamlVersion, got %q", s.OCamlVersion)
	}
}

func TestSaveState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := project.State{
		SwitchPath:   "/cache/oc/switches/abc123/",
		OCamlVersion: "5.2.0",
	}
	if err := project.SaveState(dir, s); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	got, err := project.LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if got.SwitchPath != s.SwitchPath {
		t.Errorf("SwitchPath: got %q, want %q", got.SwitchPath, s.SwitchPath)
	}
	if got.OCamlVersion != s.OCamlVersion {
		t.Errorf("OCamlVersion: got %q, want %q", got.OCamlVersion, s.OCamlVersion)
	}
}

func TestSaveState_CreatesOCDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := project.SaveState(dir, project.State{SwitchPath: "/tmp/sw"}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".oc")); err != nil {
		t.Errorf(".oc directory not created: %v", err)
	}
}

func TestSaveState_NoTempFilesLeft(t *testing.T) {
	dir := t.TempDir()
	if err := project.SaveState(dir, project.State{SwitchPath: "/tmp/sw"}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".oc"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestSaveState_ConcurrentWritesProduceValidFile(t *testing.T) {
	dir := t.TempDir()
	const goroutines = 50
	var wg gosync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			_ = project.SaveState(dir, project.State{
				SwitchPath:   fmt.Sprintf("/path/%d", i),
				OCamlVersion: "5.2.0",
			})
		}()
	}
	wg.Wait()
	_, err := project.LoadState(dir)
	if err != nil {
		t.Errorf("after concurrent SaveState, state.toml is not parseable: %v", err)
	}
}

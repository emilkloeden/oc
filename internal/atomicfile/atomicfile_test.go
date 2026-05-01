package atomicfile_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/internal/atomicfile"
)

// TestWrite_CleansTempFileOnRenameFailure verifies that no .tmp file is left
// behind when os.Rename fails. We force the failure by making the destination
// path a directory (Rename to an existing dir returns EISDIR on POSIX).
func TestWrite_CleansTempFileOnRenameFailure(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "target")
	if err := os.Mkdir(dest, 0755); err != nil {
		t.Fatal(err)
	}

	err := atomicfile.Write(dest, []byte("hello"), 0644)
	if err == nil {
		t.Fatal("expected error when rename target is a directory, got nil")
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind after rename failure: %s", e.Name())
		}
	}
}

func TestWrite_SuccessNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.txt")

	if err := atomicfile.Write(dest, []byte("content"), 0644); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind on success: %s", e.Name())
		}
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "content" {
		t.Errorf("got %q, want %q", data, "content")
	}
}

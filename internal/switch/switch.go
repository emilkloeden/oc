package swmgr

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// CachePathForVersion returns the content-addressed switch path for the given OCaml version.
// All projects using the same OCaml version share the same switch (dependencies accumulate via opam).
func CachePathForVersion(ocamlVersion string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	h := sha256.New()
	fmt.Fprintf(h, "ocaml=%s\n", ocamlVersion)
	hash := fmt.Sprintf("%x", h.Sum(nil))[:16]
	return filepath.Join(home, ".cache", "oc", "switches", hash), nil
}

// EnsureSymlink creates or updates the .ocaml symlink in projectDir to point at target.
// There is an inherent TOCTOU race between Lstat, Remove, and Symlink: a concurrent
// process could modify the symlink in that window. This is accepted as low risk because
// oc is an interactive CLI, not a concurrent daemon. The failure mode is a benign
// "symlink already exists" error on the second invocation, not data loss.
func EnsureSymlink(projectDir, target string) error {
	link := filepath.Join(projectDir, ".ocaml")

	info, err := os.Lstat(link)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Symlink(target, link); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
			return nil
		}
		return fmt.Errorf("stat .ocaml: %w", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf(".ocaml exists as a %s; remove it manually to allow oc to manage the switch symlink",
			fileTypeDescription(info.Mode()))
	}

	existing, err := os.Readlink(link)
	if err != nil {
		return fmt.Errorf("readlink .ocaml: %w", err)
	}
	if existing == target {
		return nil
	}
	if err := os.Remove(link); err != nil {
		return fmt.Errorf("remove stale symlink: %w", err)
	}
	return os.Symlink(target, link)
}

func fileTypeDescription(mode os.FileMode) string {
	switch {
	case mode.IsDir():
		return "directory"
	case mode&os.ModeSymlink != 0:
		return "symlink"
	default:
		return "regular file"
	}
}

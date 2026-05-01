package swmgr

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// CachePathForVersion returns the content-addressed switch path for the given OCaml version.
// All projects using the same OCaml version share the same switch (dependencies accumulate via opam).
func CachePathForVersion(ocamlVersion string) string {
	home, _ := os.UserHomeDir()
	h := sha256.New()
	fmt.Fprintf(h, "ocaml=%s\n", ocamlVersion)
	hash := fmt.Sprintf("%x", h.Sum(nil))[:16]
	return filepath.Join(home, ".cache", "oc", "switches", hash)
}

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

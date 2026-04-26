package swmgr

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/emilkloeden/oc/internal/project"
)

func Hash(lock *project.Lock) string {
	pkgs := make([]project.Package, len(lock.Packages))
	copy(pkgs, lock.Packages)
	sort.Slice(pkgs, func(i, j int) bool {
		if pkgs[i].Name != pkgs[j].Name {
			return pkgs[i].Name < pkgs[j].Name
		}
		return pkgs[i].Version < pkgs[j].Version
	})

	h := sha256.New()
	fmt.Fprintf(h, "ocaml=%s\n", lock.OCaml.Version)
	for _, p := range pkgs {
		fmt.Fprintf(h, "%s=%s\n", p.Name, p.Version)
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func CachePath(lock *project.Lock) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "oc", "switches", Hash(lock))
}

func EnsureSymlink(projectDir, target string) error {
	link := filepath.Join(projectDir, ".ocaml")

	if existing, err := os.Readlink(link); err == nil {
		if existing == target {
			return nil
		}
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("remove stale symlink: %w", err)
		}
	}

	return os.Symlink(target, link)
}

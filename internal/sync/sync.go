package sync

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/emilkloeden/oc/internal/exec"
	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	swmgr "github.com/emilkloeden/oc/internal/switch"
)

const defaultOCamlVersion = "5.2.0"

type OpamRunner interface {
	SwitchExists(path string) bool
	CreateSwitch(path, ocamlVersion string) error
	InstallDeps(dir, switchPath string) error
	ListInstalled(switchPath string) ([]project.Package, error)
}

type realRunner struct{}

func (r *realRunner) SwitchExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (r *realRunner) CreateSwitch(path, ocamlVersion string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create switch directory: %w", err)
	}
	fmt.Printf("Creating OCaml %s switch (this may take a minute)...\n", ocamlVersion)
	return exec.Run("opam", []string{
		"switch", "create", path, ocamlVersion, "--yes",
	}, exec.Options{})
}

func (r *realRunner) InstallDeps(dir, switchPath string) error {
	return exec.Run("opam", []string{
		"install", ".", "--deps-only", "--yes",
		"--switch", switchPath,
	}, exec.Options{Dir: dir})
}

func (r *realRunner) ListInstalled(switchPath string) ([]project.Package, error) {
	var buf bytes.Buffer
	err := exec.Run("opam", []string{
		"list", "--installed", "--short",
		"--columns=name,version",
		"--switch", switchPath,
	}, exec.Options{Stdout: &buf})
	if err != nil {
		return nil, err
	}

	var pkgs []project.Package
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkgs = append(pkgs, project.Package{Name: parts[0], Version: parts[1]})
		}
	}
	return pkgs, nil
}

// Ensure is the public entry point using the real opam runner.
// It reads the OCaml version from oc.lock, defaulting to 5.2.0 if absent.
func Ensure(dir string) error {
	if err := opam.CheckOpam(); err != nil {
		return err
	}
	lock, err := project.LoadLock(dir)
	if err != nil {
		return fmt.Errorf("load lockfile: %w", err)
	}
	ocamlVersion := lock.OCaml.Version
	if ocamlVersion == "" {
		ocamlVersion = defaultOCamlVersion
	}
	return EnsureWith(dir, ocamlVersion, &realRunner{})
}

// EnsureWith allows injection of a custom runner (used in tests).
func EnsureWith(dir string, ocamlVersion string, runner OpamRunner) error {
	lock, err := project.LoadLock(dir)
	if err != nil {
		return fmt.Errorf("load lockfile: %w", err)
	}
	// Detect OCaml version change — stale switch path must be discarded.
	if lock.OCaml.Version != "" && lock.OCaml.Version != ocamlVersion {
		lock.SwitchPath = ""
	}
	lock.OCaml.Version = ocamlVersion

	// Use the stored switch path if present and the switch still exists there.
	// This keeps the path stable even after the lock is populated with packages
	// (which would change the content-addressed hash).
	switchPath := lock.SwitchPath
	if switchPath == "" || !runner.SwitchExists(switchPath) {
		switchPath = swmgr.CachePath(lock)
	}

	if !runner.SwitchExists(switchPath) {
		if err := runner.CreateSwitch(switchPath, ocamlVersion); err != nil {
			return fmt.Errorf("create switch: %w", err)
		}
	}

	if err := swmgr.EnsureSymlink(dir, switchPath); err != nil {
		return fmt.Errorf("link switch: %w", err)
	}

	if err := runner.InstallDeps(dir, switchPath); err != nil {
		return fmt.Errorf("install deps: %w", err)
	}

	pkgs, err := runner.ListInstalled(switchPath)
	if err != nil {
		return fmt.Errorf("list installed: %w", err)
	}

	lock.SwitchPath = switchPath
	lock.Packages = pkgs
	if err := project.SaveLock(dir, lock); err != nil {
		return fmt.Errorf("save lockfile: %w", err)
	}

	return nil
}

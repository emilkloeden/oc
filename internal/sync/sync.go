package sync

import (
	"fmt"
	"os"

	"github.com/emilkloeden/oc/internal/defaults"
	"github.com/emilkloeden/oc/internal/exec"
	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	swmgr "github.com/emilkloeden/oc/internal/switch"
)

type OpamRunner interface {
	SwitchExists(path string) bool
	CreateSwitch(path, ocamlVersion string) error
	InstallDeps(dir, switchPath string) error
	LockDeps(dir string) error
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

func (r *realRunner) LockDeps(dir string) error {
	return exec.Run("opam", []string{"lock", "."}, exec.Options{Dir: dir})
}

// Ensure is the public entry point using the real opam runner.
// It reads the OCaml version from the .opam file, defaulting to 5.2.0 if absent.
func Ensure(dir string) error {
	if err := opam.CheckOpam(); err != nil {
		return err
	}
	ocamlVersion, err := opam.ReadOCamlVersion(dir)
	if err != nil {
		ocamlVersion = defaults.DefaultOCamlVersion
	}
	return EnsureWith(dir, ocamlVersion, &realRunner{})
}

// EnsureWith allows injection of a custom runner (used in tests).
func EnsureWith(dir string, ocamlVersion string, runner OpamRunner) error {
	state, err := project.LoadState(dir)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	// Detect OCaml version change — discard stale switch path.
	if state.OCamlVersion != "" && state.OCamlVersion != ocamlVersion {
		state.SwitchPath = ""
	}
	state.OCamlVersion = ocamlVersion

	switchPath := state.SwitchPath
	if switchPath == "" || !runner.SwitchExists(switchPath) {
		switchPath = swmgr.CachePathForVersion(ocamlVersion)
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

	state.SwitchPath = switchPath
	if err := project.SaveState(dir, state); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	// Generate *.opam.locked after successful install. Failure is non-fatal.
	if err := runner.LockDeps(dir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: opam lock failed: %v\n", err)
	}

	return nil
}

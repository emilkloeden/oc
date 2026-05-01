package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	CloneSwitch(source, dest string) error
	GetSwitchOCamlVersion(switchPath string) (string, error)
	ListCachedSwitches() ([]string, error)
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

func (r *realRunner) CloneSwitch(source, dest string) error {
	fmt.Printf("Cloning existing switch as base (this may take a moment)...\n")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("create switches dir: %w", err)
	}
	return exec.Run("opam", []string{
		"switch", "copy", source, dest,
	}, exec.Options{})
}

func (r *realRunner) GetSwitchOCamlVersion(switchPath string) (string, error) {
	out, err := exec.Output("opam", []string{
		"list", "ocaml", "--installed", "--short", "--columns=version",
		"--switch", switchPath,
	}, exec.Options{})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (r *realRunner) ListCachedSwitches() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determine home directory: %w", err)
	}
	return swmgr.ListCachedSwitches(filepath.Join(home, ".cache", "oc", "switches"))
}

// findBestBase returns the path of the best existing switch to clone from for
// targetVersion. It picks the first compatible switch (same OCaml version) that
// is not the target path itself. Returns "" when no suitable base exists.
func findBestBase(targetPath, ocamlVersion string, runner OpamRunner) string {
	candidates, err := runner.ListCachedSwitches()
	if err != nil || len(candidates) == 0 {
		return ""
	}
	for _, p := range candidates {
		if p == targetPath || !runner.SwitchExists(p) {
			continue
		}
		v, err := runner.GetSwitchOCamlVersion(p)
		if err != nil {
			continue
		}
		if v == ocamlVersion {
			return p
		}
	}
	return ""
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
		var err error
		switchPath, err = swmgr.CachePathForProject(dir, ocamlVersion)
		if err != nil {
			return fmt.Errorf("compute switch path: %w", err)
		}
	}

	if !runner.SwitchExists(switchPath) {
		base := findBestBase(switchPath, ocamlVersion, runner)
		if base != "" {
			if err := runner.CloneSwitch(base, switchPath); err != nil {
				// Clone failed — fall back to a clean create.
				fmt.Fprintf(os.Stderr, "warning: switch clone failed, creating from scratch: %v\n", err)
				if err2 := runner.CreateSwitch(switchPath, ocamlVersion); err2 != nil {
					return fmt.Errorf("create switch: %w", err2)
				}
			}
		} else {
			if err := runner.CreateSwitch(switchPath, ocamlVersion); err != nil {
				return fmt.Errorf("create switch: %w", err)
			}
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

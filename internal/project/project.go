package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const stateDir = ".oc"
const stateFile = "state.toml"

// State holds machine-local oc state. It is stored in .oc/state.toml and never committed.
type State struct {
	SwitchPath   string `toml:"switch_path"`
	OCamlVersion string `toml:"ocaml_version"`
}

// LoadState reads .oc/state.toml. A missing file returns an empty State with no error.
func LoadState(dir string) (State, error) {
	path := filepath.Join(dir, stateDir, stateFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}
	var s State
	if _, err := toml.Decode(string(data), &s); err != nil {
		return State{}, fmt.Errorf("parse .oc/state.toml: %w", err)
	}
	return s, nil
}

// SaveState atomically writes .oc/state.toml, creating .oc/ if needed.
func SaveState(dir string, s State) error {
	ocDir := filepath.Join(dir, stateDir)
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(ocDir, stateFile)
	tmp, err := os.CreateTemp(ocDir, ".state.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if err := toml.NewEncoder(tmp).Encode(s); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

// Dep represents a package name and its version constraint as parsed from CLI arguments.
type Dep struct {
	Name       string
	Constraint string
}

// ParseConstraintParts splits a constraint like ">=5.0.0" into op=">=", ver="5.0.0".
// Returns op="" if no recognised operator prefix is found.
func ParseConstraintParts(c string) (op, ver string) {
	for _, prefix := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(c, prefix) {
			return prefix, strings.TrimSpace(c[len(prefix):])
		}
	}
	return "", c
}

package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const lockFile = "oc.lock"

type OCamlMeta struct {
	Version string `toml:"version"`
}

type Package struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

// Dep represents a package name and its version constraint as parsed from CLI arguments.
type Dep struct {
	Name       string
	Constraint string
}

type Lock struct {
	OCaml      OCamlMeta `toml:"ocaml"`
	SwitchPath string    `toml:"switch_path,omitempty"`
	Packages   []Package `toml:"package"`
}

func LoadLock(dir string) (*Lock, error) {
	path := filepath.Join(dir, lockFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Lock{}, nil
	}
	if err != nil {
		return nil, err
	}

	var lock Lock
	if _, err := toml.Decode(string(data), &lock); err != nil {
		return nil, fmt.Errorf("parse oc.lock: %w", err)
	}
	return &lock, nil
}

func SaveLock(dir string, lock *Lock) error {
	path := filepath.Join(dir, lockFile)

	tmp, err := os.CreateTemp(dir, ".oc.lock.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if err := toml.NewEncoder(tmp).Encode(lock); err != nil {
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

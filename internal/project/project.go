package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const configFile = "oc.toml"
const lockFile = "oc.lock"

type ProjectMeta struct {
	Name       string   `toml:"name"`
	Version    string   `toml:"version"`
	Synopsis   string   `toml:"synopsis,omitempty"`
	Maintainer string   `toml:"maintainer,omitempty"`
	Authors    []string `toml:"authors,omitempty"`
	Homepage   string   `toml:"homepage,omitempty"`
	BugReports string   `toml:"bug-reports,omitempty"`
	License    string   `toml:"license,omitempty"`
}

type OCamlMeta struct {
	Version string `toml:"version"`
}

type Config struct {
	Project         ProjectMeta       `toml:"project"`
	OCaml           OCamlMeta         `toml:"ocaml"`
	Dependencies    map[string]string `toml:"dependencies"`
	DevDependencies map[string]string `toml:"dev-dependencies"`
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

func LoadConfig(dir string) (*Config, error) {
	path := filepath.Join(dir, configFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("oc.toml not found in %s", dir)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parse oc.toml: %w", err)
	}

	if cfg.Project.Name == "" {
		return nil, fmt.Errorf("oc.toml: [project] name is required")
	}

	if cfg.Dependencies == nil {
		cfg.Dependencies = map[string]string{}
	}
	if cfg.DevDependencies == nil {
		cfg.DevDependencies = map[string]string{}
	}

	return &cfg, nil
}

func SaveConfig(dir string, cfg *Config) error {
	path := filepath.Join(dir, configFile)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return toml.NewEncoder(f).Encode(cfg)
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
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return toml.NewEncoder(f).Encode(lock)
}

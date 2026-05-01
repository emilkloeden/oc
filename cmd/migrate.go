package cmd

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/emilkloeden/oc/internal/project"
)

// oldLock mirrors the fields we need from the pre-v0.4.0 oc.lock format.
type oldLock struct {
	SwitchPath string `toml:"switch_path"`
	OCaml      struct {
		Version string `toml:"version"`
	} `toml:"ocaml"`
}

// migrateIfNeeded silently migrates an old oc.lock to .oc/state.toml on first
// run after a v0.4.0 upgrade. It is a no-op when oc.lock is absent or
// .oc/state.toml already exists.
func migrateIfNeeded(dir string) {
	lockPath := filepath.Join(dir, "oc.lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return // no oc.lock — nothing to migrate
	}

	// Skip if state already exists (migration was already done).
	if _, err := os.Stat(filepath.Join(dir, ".oc", "state.toml")); err == nil {
		return
	}

	var lock oldLock
	if _, err := toml.Decode(string(data), &lock); err != nil {
		return // corrupt lock; leave it alone
	}

	state := project.State{
		SwitchPath:   lock.SwitchPath,
		OCamlVersion: lock.OCaml.Version,
	}
	if err := project.SaveState(dir, state); err != nil {
		return
	}

	_ = os.Remove(lockPath)
}

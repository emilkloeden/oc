# CLAUDE.md

## What this project is

`oc` is a Go CLI tool that wraps [opam](https://opam.ocaml.org) and [dune](https://dune.build) to give OCaml a Cargo-like developer experience. It manages per-project opam switches transparently, generates opam files from `oc.toml`, and keeps a lockfile of resolved package versions.

## Build and test

```sh
# Build
go build ./...

# Unit tests (no opam needed)
go test ./...

# Integration tests (opam must be installed and initialised)
go test -tags integration -timeout 20m .
```

Go 1.26+ is required. `opam` and `git` must be on `PATH` at runtime but are not needed to compile.

## Architecture

```
main.go                  → entry point, delegates to cmd.Execute()
cmd/                     → cobra commands; no business logic, just routing
  new.go                 → oc new: RunNew(parent, name, lib) is exported for tests
  add.go                 → oc add: updates oc.toml + opam file, then calls sync.Ensure
  remove.go              → oc remove: inverse of add
  build.go               → oc build: sync.Ensure → opam exec dune build
  run.go                 → oc run:   sync.Ensure → opam exec dune exec ./bin/main.exe
  env.go                 → oc env: reads oc.lock, prints OCaml version + packages
  util.go                → findProjectRoot: walks up from cwd looking for oc.toml
  export_test.go         → exposes unexported helpers (printEnvInfo, findProjectRoot) for tests

internal/project/        → oc.toml and oc.lock read/write (pure, no subprocess calls)
  Config struct          → [project], [ocaml], [dependencies], [dev-dependencies]
  Lock struct            → [ocaml], switch_path, [[package]] list
  LoadConfig/SaveConfig  → TOML round-trip for oc.toml
  LoadLock/SaveLock      → TOML round-trip for oc.lock; missing lock returns empty Lock{}

internal/opam/           → generates <name>.opam from a Config (pure, no subprocess calls)
  Generate(dir, cfg)     → always emits synopsis, maintainer, ocaml constraint, dune dep

internal/switch/         → switch hashing and symlink management (package name: swmgr)
  Hash(lock)             → SHA-256 of sorted "name=version\n" lines, first 16 hex chars
  CachePath(lock)        → ~/.cache/oc/switches/<hash>
  EnsureSymlink(dir, target) → creates or updates .ocaml/ symlink in project root

internal/sync/           → orchestrates switch lifecycle and dep installation
  Ensure(dir, cfg)       → real opam runner entry point
  EnsureWith(dir, cfg, OpamRunner) → injectable for unit tests
  OpamRunner interface   → SwitchExists, CreateSwitch, InstallDeps, ListInstalled

internal/exec/           → thin subprocess wrapper
  Run(name, args, opts)  → streams stdout/stderr; opts: Dir, Env, Stdout, Stderr
  Output(name, args, opts) → captures stdout as string
```

## Key design decisions

**Switch path is stored in oc.lock** (`switch_path` field). The content-addressed hash is computed from the lock before packages are installed; after installation the lock is populated with actual package versions which would change the hash. Persisting the resolved path prevents the switch from being "lost" on the second sync call.

**opam file is always generated, never edited by users.** `oc.toml` is the single source of truth. `opam.Generate` is called whenever `oc.toml` changes (new, add, remove).

**`dune` is always included as an implicit dependency** in the generated opam file so `opam install --deps-only` always installs it.

**`internal/switch` package name is `swmgr`** (not `switch`) because `switch` is a Go keyword and cannot be used as a package name.

**`sync.EnsureWith` takes an `OpamRunner` interface** so the orchestration logic can be fully unit-tested without opam installed. The real runner calls opam; tests inject a `mockRunner`.

## File ownership

| File | Owned by | Notes |
|---|---|---|
| `oc.toml` | user | edit freely |
| `oc.lock` | `oc` | generated; commit for reproducibility |
| `<name>.opam` | `oc` | generated; do not edit |
| `dune-project` | user | scaffolded once; edit freely |
| `bin/dune`, `lib/dune` | user | scaffolded once; edit freely |
| `.ocaml/` | `oc` | symlink to switch; gitignored |

## Adding a new command

1. Create `cmd/<name>.go` with a `cobra.Command` and an `init()` that calls `rootCmd.AddCommand`.
2. Extract the business logic into a named function (e.g. `RunFoo(dir string, ...) error`) so it can be called from tests without going through cobra.
3. If the function needs to be tested from `cmd_test` (external test package), add a `var FooFn = fooFn` line to `cmd/export_test.go`.
4. Write tests in `cmd/<name>_test.go` using `package cmd_test`.
5. For logic touching opam, add an integration test in `integration_test.go` with `//go:build integration`.

## Constraint syntax in oc.toml

| oc.toml value | opam output |
|---|---|
| `"*"` | `"pkg"` (no constraint) |
| `">=5.0.0"` | `"pkg" {>= "5.0.0"}` |
| `"<=2.0"` | `"pkg" {<= "2.0"}` |
| `"=1.0.0"` | `"pkg" {= "1.0.0"}` |

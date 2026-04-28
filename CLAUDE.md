# CLAUDE.md

## What this project is

`oc` is a Go CLI tool that wraps [opam](https://opam.ocaml.org) and [dune](https://dune.build) to give OCaml a Cargo-like developer experience. It manages per-project opam switches transparently, generates opam files from `oc.toml`, and keeps a lockfile of resolved package versions.

## Development workflow

### TDD is required

Always follow the red-green cycle:
1. Write a failing test that describes the behaviour you want.
2. Confirm it fails (`go test ./...`).
3. Write the minimum implementation to make it pass.
4. Confirm it passes, then move on.

Do not write implementation code before a failing test exists.

### Build and test

```sh
# Build
go build ./...

# Unit tests (no opam needed)
go test ./...

# Integration tests (opam must be installed and initialised)
go test -tags integration -timeout 20m .
```

Go 1.22+ is required. `opam` and `git` must be on `PATH` at runtime but are not needed to compile.

### Git and GitHub workflow

**No direct commits to `main`.** All changes go through a branch and PR.

**Docs-only changes** (README, CONTRIBUTING, install.sh, CLAUDE.md, docs/):
- Branch name: `docs/<short-description>`
- Open a PR; no issue required

**Code changes** (anything in `cmd/`, `internal/`, `main.go`, tests):
1. Raise a GitHub issue first: `gh issue create --repo emilkloeden/oc --title "..." --body "..."`
2. If the change targets a specific release, assign it to the milestone: `--milestone "v0.x.0"`
3. Create a branch: `git checkout -b feat/<description>` (match the commit type)
4. Open a PR that closes the issue: include `Closes #N` in the PR body
5. Merge with squash and delete the branch: `gh pr merge <N> --squash --delete-branch`

**Milestones:** use `gh api repos/emilkloeden/oc/milestones --method POST --field title="v0.x.0"` to create milestones; assign issues to them so there is a clear record of what is targeted at each release.

### Pre-commit checks (automated)

A Claude Code hook in `.claude/settings.json` intercepts every `git commit` command and runs:

```sh
go test ./...
golangci-lint run ./...
```

The commit is blocked if either command fails. Fix the reported errors and try again.

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add oc upgrade command
fix: correct switch path on second sync
perf: cache opam list output
refactor: extract version parsing
test: add wildcard constraint case
docs: update README install steps
chore: bump BurntSushi/toml to v1.5
ci: pin golangci-lint to v2.11
```

`docs:`, `test:`, `chore:`, `ci:` are excluded from release changelogs. Use `feat:`, `fix:`, `perf:` for anything users should see. Breaking changes: `feat!:` with a `BREAKING CHANGE:` footer.

### Releases

Releases are automated via GoReleaser, triggered by pushing a `v*` tag:

```sh
# 1. Confirm CI is green on main
# 2. Tag and push
git tag v0.x.0
git push origin v0.x.0
```

GoReleaser builds binaries for Linux/macOS/Windows (amd64 + arm64), creates a GitHub Release with archives and checksums, and generates a changelog from conventional commit prefixes. The version is baked into the binary at build time — the tag is the source of truth. Use semver (`v0.1.0`, `v0.2.0`, `v1.0.0`).

### Documentation standards

- Write "dependencies" in prose, never "deps"
- When naming runtime prerequisites by name in documentation, link to their install page:
  - [opam](https://opam.ocaml.org/doc/Install.html)
  - [git](https://git-scm.com)
- `go.mod` is the authoritative source for the minimum Go version; keep any version references in docs in sync with it

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

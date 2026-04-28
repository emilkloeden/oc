# Contributing to oc

## Development setup

Go 1.22+ is required. [opam](https://opam.ocaml.org/doc/Install.html) and [git](https://git-scm.com) must be on `PATH` at runtime but are not needed to compile or run unit tests.

## Workflow

This project follows a strict TDD cycle: write a failing test, confirm it fails, write the minimum implementation to pass it, confirm it passes.

## Build and test

```sh
# Build
go build ./...

# Unit tests (no opam needed)
go test ./...

# Integration tests (opam must be installed and initialised)
opam init --bare
go test -tags integration -timeout 20m .
```

## Commit style

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add oc upgrade command
fix: correct switch path on second sync
chore: bump BurntSushi/toml to v1.5
```

`feat:` and `fix:` appear in the release changelog; `chore:`, `docs:`, `test:`, `ci:` do not.

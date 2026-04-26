# oc

A Cargo-like developer experience for OCaml. One command surface, zero manual switch management — `oc` orchestrates [opam](https://opam.ocaml.org) and [dune](https://dune.build) behind the scenes so you never have to.

> **Goal:** "I didn't learn opam. I didn't learn dune. I just built something."

## Installation

**Prerequisites:** [opam](https://opam.ocaml.org/doc/Install.html) and [git](https://git-scm.com) must be on your `PATH`. Go is only needed to build `oc` itself.

```sh
go install github.com/emilkloeden/oc@latest
```

Or build from source:

```sh
git clone https://github.com/emilkloeden/oc
cd oc
go build -o ~/.local/bin/oc .
```

## Quick start

```sh
oc new my_app
cd my_app
oc run
```

That's it. `oc` creates a local OCaml environment, installs dependencies, and runs your program — no `opam switch create`, no `eval $(opam env)`, no manual `dune` invocations.

## Commands

### `oc new <name>`

Create a new project.

```sh
oc new my_app          # binary project (default)
oc new my_lib --lib    # library project
```

Creates:

```
my_app/
├── .git/
├── .gitignore
├── oc.toml            ← your config (edit this)
├── my_app.opam        ← generated, do not edit
├── dune-project
└── bin/
    ├── dune
    └── main.ml
```

### `oc build`

Sync dependencies and build.

```sh
oc build
```

Ensures the project switch exists, installs any missing deps, then runs `dune build`.

### `oc run`

Build and run the project binary.

```sh
oc run
```

### `oc add <package> [constraint]`

Add a dependency.

```sh
oc add cohttp            # any version
oc add cohttp ">=5.0.0"  # with constraint
oc add alcotest --dev    # dev-only dependency
```

Updates `oc.toml`, regenerates `my_app.opam`, installs the package into the project switch, and writes the resolved versions to `oc.lock`.

### `oc remove <package>`

Remove a dependency.

```sh
oc remove cohttp
```

### `oc env`

Show the project environment.

```sh
oc env
# ocaml    5.2.0
# packages (3 installed)
#   dune                           3.22.2
#   ocaml                          5.2.0
#   stringext                      1.6.0
# switch   /Users/you/.cache/oc/switches/abc123
```

## Project configuration

Edit `oc.toml` to configure your project:

```toml
[project]
name = "my_app"
version = "0.1.0"
synopsis = "My great OCaml application"
maintainer = "Alice <alice@example.com>"

[ocaml]
version = "5.2.0"

[dependencies]
cohttp = ">=5.0.0"
lwt = "*"

[dev-dependencies]
alcotest = "*"
```

**Do not edit `my_app.opam`** — it is generated from `oc.toml` and overwritten on every change.

## How it works

`oc` is a thin orchestration layer over opam and dune. It never reimplements package solving or compilation.

**Switch strategy:** each project gets an isolated opam switch stored in a global content-addressed cache at `~/.cache/oc/switches/<hash>/`. The project root contains a `.ocaml/` symlink pointing to it. Two projects with the same OCaml version and dependency set share one switch automatically.

**Lockfile:** `oc.lock` records the exact installed versions and the resolved switch path. Commit it for reproducible builds.

```
~/.cache/oc/
└── switches/
    └── a5ea70a0aa46624e/   ← shared by any project with this exact dep set
```

## Running integration tests

Unit tests (`go test ./...`) run without opam. Integration tests require opam to be initialised:

```sh
opam init --bare
go test -tags integration -timeout 20m .
```

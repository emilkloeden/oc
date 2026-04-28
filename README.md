# oc

A Cargo-like developer experience for OCaml. One command surface, zero manual switch management — `oc` orchestrates [opam](https://opam.ocaml.org) and [dune](https://dune.build) behind the scenes so you never have to.

> **Goal:** "I didn't learn opam. I didn't learn dune. I just built something."

![oc demo](docs/demo.gif)

## Installation

**Prerequisites:** [opam](https://opam.ocaml.org/doc/Install.html) and [git](https://git-scm.com) must be on your `PATH`.

### macOS / Linux (recommended)

```sh
curl -sSfL https://raw.githubusercontent.com/emilkloeden/oc/main/install.sh | sh
```

Detects your OS and architecture, downloads the correct binary from the latest release, and installs to `/usr/local/bin` (uses `sudo` if needed). To install elsewhere:

```sh
curl -sSfL https://raw.githubusercontent.com/emilkloeden/oc/main/install.sh | INSTALL_DIR=~/.local/bin sh
```

### Manual

Download the binary for your platform from the [latest release](https://github.com/emilkloeden/oc/releases/latest), extract it, and place `oc` somewhere on your `PATH`.

### Build from source

Requires Go 1.22+.

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

Ensures the project switch exists, installs any missing dependencies, then runs `dune build`.

### `oc run`

Build and run the project binary.

```sh
oc run
```

### `oc add <package> [constraint] ...`

Add one or more dependencies.

```sh
oc add cohttp                        # any version
oc add cohttp ">=5.0.0"              # with constraint
oc add cohttp-lwt-unix yojson lwt    # multiple packages at once
oc add alcotest --dev                # dev-only dependency
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

## Why oc?

Raw opam is powerful but manual. A typical project setup looks like:

```sh
opam switch create my_app 5.2.0
eval $(opam env)
opam install dune cohttp-lwt-unix
dune init project my_app
```

And that's before you've written a line of OCaml. Switch to a different project and repeat — or remember which switch you're on.

With `oc`:

```sh
oc new my_app
cd my_app
oc run
```

`oc` handles switch creation, environment activation, and dependency installation automatically. It doesn't replace opam or dune — it orchestrates them so you don't have to.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, the TDD workflow this project follows, and how to run integration tests.

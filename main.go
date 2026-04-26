package main

import "github.com/emilkloeden/oc/cmd"

// version is set at build time by GoReleaser via -ldflags "-X main.version=<tag>".
// It defaults to "dev" for local builds.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}

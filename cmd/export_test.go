package cmd

import (
	"io"

	"github.com/emilkloeden/oc/internal/project"
)

// Exported for testing only.
var PrintEnvInfo = func(w io.Writer, lock *project.Lock) { printEnvInfo(w, lock) }
var FindProjectRoot = findProjectRoot
var BuildRunArgs = buildRunArgs
var ParseAddArgs = parseAddArgs

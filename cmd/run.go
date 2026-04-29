package cmd

import (
	"fmt"

	"github.com/emilkloeden/oc/internal/exec"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
	"github.com/spf13/cobra"
)

func buildRunArgs(switchPath string, extraArgs ...string) []string {
	args := []string{"exec", "--switch", switchPath, "--", "dune", "exec", "./bin/main.exe"}
	if len(extraArgs) > 0 {
		args = append(args, "--")
		args = append(args, extraArgs...)
	}
	return args
}

var runCmd = &cobra.Command{
	Use:                "run [-- args...]",
	Short:              "Build and run the project",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := projectRoot()
		if err != nil {
			return err
		}
		return runRun(dir, args)
	},
}

// runRun performs the build-and-run for the given project directory.
func runRun(dir string, args []string) error {
	if err := sync.Ensure(dir); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	lock, err := project.LoadLock(dir)
	if err != nil {
		return fmt.Errorf("load lockfile: %w", err)
	}
	switchPath := lock.SwitchPath

	fmt.Println("Building and running...")
	return exec.Run("opam", buildRunArgs(switchPath, args...), exec.Options{Dir: dir})
}

func init() {
	rootCmd.AddCommand(runCmd)
}

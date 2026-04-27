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

		cfg, err := project.LoadConfig(dir)
		if err != nil {
			return err
		}

		if err := sync.Ensure(dir, cfg); err != nil {
			return err
		}

		lock, _ := project.LoadLock(dir)
		switchPath := lock.SwitchPath

		fmt.Println("Building and running...")
		return exec.Run("opam", buildRunArgs(switchPath, args...), exec.Options{Dir: dir})
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

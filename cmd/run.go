package cmd

import (
	"fmt"

	"github.com/emilkloeden/oc/internal/exec"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and run the project",
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
		return exec.Run("opam", []string{
			"exec", "--switch", switchPath, "--", "dune", "exec", "./bin/main.exe",
		}, exec.Options{Dir: dir})
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

package cmd

import (
	"fmt"

	"github.com/emilkloeden/oc/internal/exec"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := projectRoot()
		if err != nil {
			return err
		}
		return runBuild(dir)
	},
}

// runBuild performs the build for the given project directory.
func runBuild(dir string) error {
	if err := sync.Ensure(dir); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	state, err := project.LoadState(dir)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	fmt.Println("Building...")
	return exec.Run("opam", []string{
		"exec", "--switch", state.SwitchPath, "--", "dune", "build",
	}, exec.Options{Dir: dir})
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

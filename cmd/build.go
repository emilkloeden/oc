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
	cfg, err := project.LoadConfig(dir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := sync.Ensure(dir, cfg); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	lock, err := project.LoadLock(dir)
	if err != nil {
		return fmt.Errorf("load lockfile: %w", err)
	}
	switchPath := lock.SwitchPath

	fmt.Println("Building...")
	return exec.Run("opam", []string{
		"exec", "--switch", switchPath, "--", "dune", "build",
	}, exec.Options{Dir: dir})
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

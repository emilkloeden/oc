package cmd

import (
	"fmt"

	"github.com/emilkloeden/oc/internal/exec"
	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Remove a dependency from the project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		dir, err := projectRoot()
		if err != nil {
			return err
		}

		cfg, err := project.LoadConfig(dir)
		if err != nil {
			return err
		}

		_, inDeps := cfg.Dependencies[pkg]
		_, inDev := cfg.DevDependencies[pkg]
		if !inDeps && !inDev {
			return fmt.Errorf("%q is not a dependency of this project", pkg)
		}

		delete(cfg.Dependencies, pkg)
		delete(cfg.DevDependencies, pkg)

		if err := project.SaveConfig(dir, cfg); err != nil {
			return fmt.Errorf("save oc.toml: %w", err)
		}
		if err := opam.Generate(dir, cfg); err != nil {
			return fmt.Errorf("regenerate opam file: %w", err)
		}

		fmt.Printf("Removing %s...\n", pkg)
		_ = exec.Run("opam", []string{"remove", pkg, "--yes"}, exec.Options{Dir: dir})

		fmt.Printf("Removed %q from dependencies.\n", pkg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

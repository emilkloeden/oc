package cmd

import (
	"fmt"

	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
	"github.com/spf13/cobra"
)

var addDev bool

var addCmd = &cobra.Command{
	Use:   "add <package> [constraint]",
	Short: "Add a dependency to the project",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]
		constraint := "*"
		if len(args) == 2 {
			constraint = args[1]
		}

		dir, err := projectRoot()
		if err != nil {
			return err
		}

		cfg, err := project.LoadConfig(dir)
		if err != nil {
			return err
		}

		if addDev {
			cfg.DevDependencies[pkg] = constraint
		} else {
			cfg.Dependencies[pkg] = constraint
		}

		if err := project.SaveConfig(dir, cfg); err != nil {
			return fmt.Errorf("save oc.toml: %w", err)
		}
		if err := opam.Generate(dir, cfg); err != nil {
			return fmt.Errorf("regenerate opam file: %w", err)
		}

		// sync.Ensure installs deps, updates the lockfile, and ensures the switch exists.
		if err := sync.Ensure(dir, cfg); err != nil {
			return err
		}

		fmt.Printf("Added %q to dependencies.\n", pkg)
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&addDev, "dev", false, "add as a dev dependency")
	rootCmd.AddCommand(addCmd)
}

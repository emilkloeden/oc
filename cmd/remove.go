package cmd

import (
	"fmt"

	"github.com/emilkloeden/oc/internal/dune"
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
		dir, err := projectRoot()
		if err != nil {
			return err
		}
		return runRemove(dir, args[0])
	},
}

// runRemove removes a dependency from the project manifest and uninstalls it via opam.
func runRemove(dir, pkg string) error {
	pt, err := project.Detect(dir)
	if err != nil {
		return err
	}

	switch pt {
	case project.TypeDuneManaged:
		if err := dune.RemoveDep(dir, pkg); err != nil {
			return fmt.Errorf("remove %s from dune-project: %w", pkg, err)
		}
	case project.TypeHandWrittenOpam:
		path, err := opam.FindOpamFile(dir)
		if err != nil {
			return err
		}
		if err := opam.RemoveDepFromOpam(path, pkg); err != nil {
			return fmt.Errorf("remove %s from opam file: %w", pkg, err)
		}
	}

	fmt.Printf("Removing %s...\n", pkg)
	if err := exec.Run("opam", []string{"remove", pkg, "--yes"}, exec.Options{Dir: dir}); err != nil {
		return fmt.Errorf("opam remove: %w", err)
	}

	fmt.Printf("Removed %q from dependencies.\n", pkg)
	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

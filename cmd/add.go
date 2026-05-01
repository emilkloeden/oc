package cmd

import (
	"fmt"
	"strings"

	"github.com/emilkloeden/oc/internal/dune"
	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <package> [constraint] [<package> [constraint]]...",
	Short: "Add one or more dependencies to the project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := parseAddArgs(args)
		if err != nil {
			return err
		}

		dir, err := projectRoot()
		if err != nil {
			return err
		}

		if err := runAdd(dir, deps, sync.Ensure); err != nil {
			return err
		}

		for _, d := range deps {
			fmt.Printf("Added %q to dependencies.\n", d.Name)
		}
		return nil
	},
}

// runAdd adds deps to the project manifest and syncs.
func runAdd(dir string, deps []project.Dep, syncFn func(string) error) error {
	pt, err := project.Detect(dir)
	if err != nil {
		return err
	}

	for _, d := range deps {
		switch pt {
		case project.TypeDuneManaged:
			if err := dune.AddDep(dir, d.Name, d.Constraint); err != nil {
				return fmt.Errorf("add %s to dune-project: %w", d.Name, err)
			}
		case project.TypeHandWrittenOpam:
			path, err := opam.FindOpamFile(dir)
			if err != nil {
				return err
			}
			if err := opam.AddDepToOpam(path, d.Name, d.Constraint); err != nil {
				return fmt.Errorf("add %s to opam file: %w", d.Name, err)
			}
		}
	}

	if err := syncFn(dir); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	return nil
}

// isConstraint reports whether an argument looks like a version constraint rather than
// a package name. Constraints start with >=, <=, =, ~, or *.
func isConstraint(s string) bool {
	return strings.HasPrefix(s, ">=") ||
		strings.HasPrefix(s, "<=") ||
		strings.HasPrefix(s, "=") ||
		strings.HasPrefix(s, "~") ||
		s == "*"
}

// parseAddArgs parses the positional arguments for "oc add" into a slice of Dep values.
// The rule is: an argument that looks like a constraint (starts with >=, <=, =, ~, or is *)
// is attached to the preceding package. Otherwise it starts a new package entry.
func parseAddArgs(args []string) ([]project.Dep, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("at least one package name is required")
	}

	var deps []project.Dep
	for _, arg := range args {
		if isConstraint(arg) {
			if len(deps) == 0 {
				return nil, fmt.Errorf("constraint %q given before any package name", arg)
			}
			deps[len(deps)-1].Constraint = arg
		} else {
			deps = append(deps, project.Dep{Name: arg, Constraint: "*"})
		}
	}
	return deps, nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}

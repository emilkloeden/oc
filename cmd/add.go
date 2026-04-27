package cmd

import (
	"fmt"
	"strings"

	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/emilkloeden/oc/internal/sync"
	"github.com/spf13/cobra"
)

var addDev bool

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

		cfg, err := project.LoadConfig(dir)
		if err != nil {
			return err
		}

		for _, d := range deps {
			if addDev {
				cfg.DevDependencies[d.Name] = d.Constraint
			} else {
				cfg.Dependencies[d.Name] = d.Constraint
			}
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

		for _, d := range deps {
			fmt.Printf("Added %q to dependencies.\n", d.Name)
		}
		return nil
	},
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
	addCmd.Flags().BoolVar(&addDev, "dev", false, "add as a dev dependency")
	rootCmd.AddCommand(addCmd)
}

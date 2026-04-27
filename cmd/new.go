package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emilkloeden/oc/internal/dune"
	"github.com/emilkloeden/oc/internal/exec"
	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/spf13/cobra"
)

var newLib bool

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new OCaml project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		return RunNew(cwd, args[0], newLib)
	},
}

// RunNew creates a new project under parent/name. Extracted for testability.
func RunNew(parent, name string, lib bool) error {
	dir := filepath.Join(parent, name)

	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("directory %q already exists", dir)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	maintainer := gitMaintainer()
	authors := []string{"Your Name <you@example.com>"}
	if maintainer != "" {
		authors = []string{maintainer}
	}

	cfg := &project.Config{
		Project: project.ProjectMeta{
			Name:       name,
			Version:    "0.1.0",
			Maintainer: maintainer,
			Authors:    authors,
			Homepage:   "https://github.com/you/" + name,
			BugReports: "https://github.com/you/" + name + "/issues",
			License:    "MIT",
		},
		OCaml:           project.OCamlMeta{Version: "5.2.0"},
		Dependencies:    map[string]string{},
		DevDependencies: map[string]string{},
	}

	if err := project.SaveConfig(dir, cfg); err != nil {
		return fmt.Errorf("write oc.toml: %w", err)
	}
	if err := opam.Generate(dir, cfg); err != nil {
		return fmt.Errorf("generate opam file: %w", err)
	}

	if lib {
		if err := dune.ScaffoldLib(dir, name); err != nil {
			return fmt.Errorf("scaffold lib: %w", err)
		}
	} else {
		if err := dune.ScaffoldBin(dir, name); err != nil {
			return fmt.Errorf("scaffold bin: %w", err)
		}
	}

	if err := writeGitignore(dir); err != nil {
		return err
	}
	initGit(dir)

	fmt.Printf("Created %q. Run:\n  cd %s\n  oc run\n", name, name)
	return nil
}

func writeGitignore(dir string) error {
	content := ".ocaml/\n_build/\n*.install\n"
	return os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644)
}

func gitMaintainer() string {
	name, _ := exec.Output("git", []string{"config", "user.name"}, exec.Options{})
	email, _ := exec.Output("git", []string{"config", "user.email"}, exec.Options{})
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name != "" && email != "" {
		return name + " <" + email + ">"
	}
	if email != "" {
		return email
	}
	return ""
}

func initGit(dir string) {
	// Non-fatal: git may not be installed, and that's okay.
	_ = exec.Run("git", []string{"init", dir}, exec.Options{})
	_ = exec.Run("git", []string{"-C", dir, "add", "."}, exec.Options{})
}

func init() {
	newCmd.Flags().BoolVar(&newLib, "lib", false, "scaffold a library instead of a binary")
	rootCmd.AddCommand(newCmd)
}

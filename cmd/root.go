package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

var currentVersion = "dev"

var rootCmd = &cobra.Command{
	Use:   "oc",
	Short: "A Cargo-like developer experience for OCaml",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Silently migrate oc.lock → .oc/state.toml on first run after upgrade.
		if dir, err := projectRoot(); err == nil {
			migrateIfNeeded(dir)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// SetVersion sets the version string reported by --version.
func SetVersion(v string) {
	currentVersion = v
	rootCmd.Version = v
}

// Version returns the current version string.
func Version() string {
	return currentVersion
}

// SetOutput redirects cobra output (used in tests). Pass nil to reset to stdout.
func SetOutput(w io.Writer) {
	if w == nil {
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetErr(os.Stderr)
	} else {
		rootCmd.SetOut(w)
		rootCmd.SetErr(w)
	}
}

// RunWithArgs executes the root command with the given arguments (used in tests).
func RunWithArgs(args []string) {
	rootCmd.SetArgs(args)
	rootCmd.Execute() //nolint:errcheck
	rootCmd.SetArgs(nil)
}

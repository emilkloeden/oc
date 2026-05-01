package cmd

import (
	"fmt"
	"strings"

	"github.com/emilkloeden/oc/internal/opam"
	"github.com/emilkloeden/oc/internal/project"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show the project environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := projectRoot()
		if err != nil {
			return err
		}

		ocamlVersion, err := opam.ReadOCamlVersion(dir)
		if err != nil {
			ocamlVersion = "(unknown)"
		}

		state, err := project.LoadState(dir)
		if err != nil {
			return err
		}

		fmt.Print(formatEnvOutput(ocamlVersion, state.SwitchPath))
		return nil
	},
}

func formatEnvOutput(ocamlVersion, switchPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ocaml    %s\n", ocamlVersion)
	if switchPath != "" {
		fmt.Fprintf(&b, "switch   %s\n", switchPath)
	} else {
		fmt.Fprintf(&b, "switch   (not yet initialised — run 'oc build')\n")
	}
	return b.String()
}

func init() {
	rootCmd.AddCommand(envCmd)
}

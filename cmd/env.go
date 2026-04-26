package cmd

import (
	"fmt"
	"io"
	"os"

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

		lock, err := project.LoadLock(dir)
		if err != nil {
			return err
		}

		printEnvInfo(os.Stdout, lock)

		if lock.SwitchPath != "" {
			fmt.Printf("switch   %s\n", lock.SwitchPath)
		} else {
			fmt.Printf("switch   (not yet initialised — run 'oc build')\n")
		}
		return nil
	},
}

func printEnvInfo(w io.Writer, lock *project.Lock) {
	fmt.Fprintf(w, "ocaml    %s\n", lock.OCaml.Version)

	if len(lock.Packages) == 0 {
		fmt.Fprintf(w, "packages (no packages installed)\n")
		return
	}

	fmt.Fprintf(w, "packages (%d installed)\n", len(lock.Packages))
	for _, p := range lock.Packages {
		fmt.Fprintf(w, "  %-30s %s\n", p.Name, p.Version)
	}
}

func init() {
	rootCmd.AddCommand(envCmd)
}

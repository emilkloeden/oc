package dune

import (
	"fmt"
	"os"
	"path/filepath"
)

const duneVersion = "3.0"

func ScaffoldBin(dir, name string) error {
	if err := writeIfAbsent(filepath.Join(dir, "dune-project"), duneProject(name)); err != nil {
		return err
	}
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	if err := writeIfAbsent(filepath.Join(binDir, "dune"), binDune(name)); err != nil {
		return err
	}
	return writeIfAbsent(filepath.Join(binDir, "main.ml"), mainML(name))
}

func ScaffoldLib(dir, name string) error {
	if err := writeIfAbsent(filepath.Join(dir, "dune-project"), duneProject(name)); err != nil {
		return err
	}
	libDir := filepath.Join(dir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return err
	}
	if err := writeIfAbsent(filepath.Join(libDir, "dune"), libDune(name)); err != nil {
		return err
	}
	return writeIfAbsent(filepath.Join(libDir, name+".ml"), libML(name))
}

func duneProject(name string) string {
	return fmt.Sprintf("(lang dune %s)\n(name %s)\n", duneVersion, name)
}

func binDune(name string) string {
	return fmt.Sprintf("(executable\n (name main)\n (public_name %s))\n", name)
}

func libDune(name string) string {
	return fmt.Sprintf("(library\n (name %s)\n (public_name %s))\n", name, name)
}

func mainML(name string) string {
	return fmt.Sprintf("let () = print_endline \"Hello from %s!\"\n", name)
}

func libML(name string) string {
	return fmt.Sprintf("let hello () = Printf.printf \"Hello from %s!\\n\"\n", name)
}

func writeIfAbsent(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists, leave it alone
	}
	return os.WriteFile(path, []byte(content), 0644)
}

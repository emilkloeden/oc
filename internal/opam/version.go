package opam

import (
	"fmt"
	osexec "os/exec"
	"strconv"
	"strings"

	"github.com/emilkloeden/oc/internal/exec"
)

// ParseOpamVersion parses a version string like "2.1.6" or "2.1" into major
// and minor integers. It returns an error if the string cannot be parsed.
func ParseOpamVersion(s string) (major, minor int, err error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("cannot parse opam version %q: expected at least major.minor", s)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse opam version %q: %w", s, err)
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse opam version %q: %w", s, err)
	}
	return major, minor, nil
}

// OpamVersionSatisfied returns true when major.minor meets the minimum
// requirement of opam >= 2.1.
func OpamVersionSatisfied(major, minor int) bool {
	return major > 2 || (major == 2 && minor >= 1)
}

// CheckOpam verifies that opam is on PATH and is version 2.1 or later.
// It returns a descriptive, actionable error if either check fails.
func CheckOpam() error {
	if _, err := osexec.LookPath("opam"); err != nil {
		return fmt.Errorf("opam not found on PATH.\nInstall it from: https://opam.ocaml.org/doc/Install.html")
	}

	out, err := exec.Output("opam", []string{"--version"}, exec.Options{})
	if err != nil {
		return fmt.Errorf("could not determine opam version: %w", err)
	}

	version := strings.TrimSpace(out)
	major, minor, err := ParseOpamVersion(version)
	if err != nil {
		return fmt.Errorf("could not parse opam version %q: %w", version, err)
	}

	if !OpamVersionSatisfied(major, minor) {
		return fmt.Errorf("opam 2.1 or later is required (found %s).\nUpgrade: https://opam.ocaml.org/doc/Install.html", version)
	}

	return nil
}

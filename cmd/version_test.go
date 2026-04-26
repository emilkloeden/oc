package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/cmd"
)

func TestVersion_DefaultIsDev(t *testing.T) {
	if cmd.Version() != "dev" {
		t.Errorf("default version: got %q want %q", cmd.Version(), "dev")
	}
}

func TestVersion_CanBeSet(t *testing.T) {
	cmd.SetVersion("1.2.3")
	defer cmd.SetVersion("dev")

	if cmd.Version() != "1.2.3" {
		t.Errorf("after SetVersion: got %q want %q", cmd.Version(), "1.2.3")
	}
}

func TestVersion_FlagPrintsVersion(t *testing.T) {
	cmd.SetVersion("0.1.0")
	defer cmd.SetVersion("dev")

	var buf bytes.Buffer
	cmd.SetOutput(&buf)
	defer cmd.SetOutput(nil)

	cmd.RunWithArgs([]string{"--version"})

	if !strings.Contains(buf.String(), "0.1.0") {
		t.Errorf("--version output missing version: %q", buf.String())
	}
}

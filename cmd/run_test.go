package cmd_test

import (
	"testing"

	"github.com/emilkloeden/oc/cmd"
)

func TestBuildRunArgs_NoExtraArgs(t *testing.T) {
	args := cmd.BuildRunArgs("/path/to/switch")
	want := []string{"exec", "--switch", "/path/to/switch", "--", "dune", "exec", "./bin/main.exe"}
	if len(args) != len(want) {
		t.Fatalf("got %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestBuildRunArgs_WithExtraArgs(t *testing.T) {
	args := cmd.BuildRunArgs("/path/to/switch", "post", "1")
	want := []string{"exec", "--switch", "/path/to/switch", "--", "dune", "exec", "./bin/main.exe", "--", "post", "1"}
	if len(args) != len(want) {
		t.Fatalf("got %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

package opam_test

import (
	"testing"

	"github.com/emilkloeden/oc/internal/opam"
)

func TestParseOpamVersion_Valid(t *testing.T) {
	major, minor, err := opam.ParseOpamVersion("2.1.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if major != 2 || minor != 1 {
		t.Errorf("expected 2.1, got %d.%d", major, minor)
	}
}

func TestParseOpamVersion_Short(t *testing.T) {
	major, minor, err := opam.ParseOpamVersion("2.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if major != 2 || minor != 1 {
		t.Errorf("expected 2.1, got %d.%d", major, minor)
	}
}

func TestParseOpamVersion_Invalid(t *testing.T) {
	_, _, err := opam.ParseOpamVersion("notaversion")
	if err == nil {
		t.Fatal("expected error for invalid version string, got nil")
	}
}

func TestOpamVersionSatisfied_ExactMinimum(t *testing.T) {
	if !opam.OpamVersionSatisfied(2, 1) {
		t.Error("2.1 should satisfy the minimum requirement")
	}
}

func TestOpamVersionSatisfied_NewerMajor(t *testing.T) {
	if !opam.OpamVersionSatisfied(3, 0) {
		t.Error("3.0 should satisfy the minimum requirement")
	}
}

func TestOpamVersionSatisfied_TooOld(t *testing.T) {
	if opam.OpamVersionSatisfied(2, 0) {
		t.Error("2.0 should NOT satisfy the minimum requirement")
	}
}

func TestOpamVersionSatisfied_OldMajor(t *testing.T) {
	if opam.OpamVersionSatisfied(1, 9) {
		t.Error("1.9 should NOT satisfy the minimum requirement")
	}
}

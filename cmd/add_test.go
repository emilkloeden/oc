package cmd_test

import (
	"testing"

	"github.com/emilkloeden/oc/cmd"
	"github.com/emilkloeden/oc/internal/project"
)

func TestParseAddArgs_SinglePackage(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []project.Dep{{Name: "pkg1", Constraint: "*"}}
	assertDeps(t, deps, want)
}

func TestParseAddArgs_SinglePackageWithConstraint(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg1", ">=1.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []project.Dep{{Name: "pkg1", Constraint: ">=1.0"}}
	assertDeps(t, deps, want)
}

func TestParseAddArgs_TwoPackagesNoConstraints(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg1", "pkg2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []project.Dep{
		{Name: "pkg1", Constraint: "*"},
		{Name: "pkg2", Constraint: "*"},
	}
	assertDeps(t, deps, want)
}

func TestParseAddArgs_PackageWithConstraintThenPackage(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg1", ">=1.0", "pkg2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []project.Dep{
		{Name: "pkg1", Constraint: ">=1.0"},
		{Name: "pkg2", Constraint: "*"},
	}
	assertDeps(t, deps, want)
}

func TestParseAddArgs_TwoPackagesEachWithConstraint(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg1", ">=1.0", "pkg2", "<=2.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []project.Dep{
		{Name: "pkg1", Constraint: ">=1.0"},
		{Name: "pkg2", Constraint: "<=2.0"},
	}
	assertDeps(t, deps, want)
}

func TestParseAddArgs_Empty_ReturnsError(t *testing.T) {
	_, err := cmd.ParseAddArgs([]string{})
	if err == nil {
		t.Fatal("expected error for empty args, got nil")
	}
}

// assertDeps checks that got matches want (order-sensitive).
func assertDeps(t *testing.T, got, want []project.Dep) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(deps) = %d, want %d; got %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Name != w.Name || got[i].Constraint != w.Constraint {
			t.Errorf("deps[%d] = %+v, want %+v", i, got[i], w)
		}
	}
}

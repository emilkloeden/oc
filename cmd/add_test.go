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

// TestParseAddArgs_InvalidConstraintPassedThrough documents that parseAddArgs does not
// validate the version portion of a constraint — it passes the token through to opam as-is.
// Validation of version strings is opam's responsibility.
func TestParseAddArgs_InvalidConstraintPassedThrough(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg", ">=invalid"})
	if err != nil {
		t.Fatalf("expected parseAddArgs to accept a syntactically-valid-but-semantically-invalid constraint, got error: %v", err)
	}
	want := []project.Dep{{Name: "pkg", Constraint: ">=invalid"}}
	assertDeps(t, deps, want)
}

// TestParseAddArgs_ConstraintOnlyNoPackage documents that a leading constraint token
// (with no preceding package name) returns an error.
func TestParseAddArgs_ConstraintOnlyNoPackage(t *testing.T) {
	_, err := cmd.ParseAddArgs([]string{">=1.0"})
	if err == nil {
		t.Fatal("expected error when constraint is given before any package name, got nil")
	}
}

// TestParseAddArgs_EmptyConstraintOperatorPassedThrough documents that a bare operator
// token (e.g. ">=") that starts with a constraint prefix is treated as a constraint
// and attached to the preceding package.
func TestParseAddArgs_EmptyConstraintOperatorPassedThrough(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg", ">="})
	if err != nil {
		t.Fatalf("expected parseAddArgs to accept a bare operator as a constraint token, got error: %v", err)
	}
	want := []project.Dep{{Name: "pkg", Constraint: ">="}}
	assertDeps(t, deps, want)
}

// TestParseAddArgs_WildcardConstraint documents that "*" is a valid constraint token
// and results in no version constraint in the generated opam file.
func TestParseAddArgs_WildcardConstraint(t *testing.T) {
	deps, err := cmd.ParseAddArgs([]string{"pkg", "*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []project.Dep{{Name: "pkg", Constraint: "*"}}
	assertDeps(t, deps, want)
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

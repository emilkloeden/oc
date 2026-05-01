package defaults_test

import (
	"strings"
	"testing"

	"github.com/emilkloeden/oc/internal/defaults"
)

func TestDefaultOCamlVersion_IsSet(t *testing.T) {
	if defaults.DefaultOCamlVersion == "" {
		t.Fatal("DefaultOCamlVersion must not be empty")
	}
}

func TestDefaultOCamlVersion_IsSemver(t *testing.T) {
	v := defaults.DefaultOCamlVersion
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		t.Errorf("DefaultOCamlVersion %q does not look like a semver (expected at least MAJOR.MINOR)", v)
	}
}

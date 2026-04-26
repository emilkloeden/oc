package exec_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emilkloeden/oc/internal/exec"
)

func TestRun_Success(t *testing.T) {
	err := exec.Run("echo", []string{"hello"}, exec.Options{})
	if err != nil {
		t.Fatalf("Run echo: %v", err)
	}
}

func TestRun_Failure(t *testing.T) {
	err := exec.Run("false", nil, exec.Options{})
	if err == nil {
		t.Fatal("expected error from 'false'")
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	err := exec.Run("no_such_command_xyz", nil, exec.Options{})
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestRun_CaptureOutput(t *testing.T) {
	var buf bytes.Buffer
	err := exec.Run("echo", []string{"captured"}, exec.Options{Stdout: &buf})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "captured") {
		t.Errorf("expected output to contain 'captured', got %q", buf.String())
	}
}

func TestRun_WithEnv(t *testing.T) {
	var buf bytes.Buffer
	err := exec.Run("sh", []string{"-c", "echo $OC_TEST_VAR"}, exec.Options{
		Stdout: &buf,
		Env:    []string{"OC_TEST_VAR=hello_from_oc"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "hello_from_oc") {
		t.Errorf("env var not passed through: %q", buf.String())
	}
}

func TestRun_WithDir(t *testing.T) {
	var buf bytes.Buffer
	err := exec.Run("pwd", nil, exec.Options{
		Stdout: &buf,
		Dir:    "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	// /tmp on macOS resolves to /private/tmp
	out := strings.TrimSpace(buf.String())
	if out != "/tmp" && out != "/private/tmp" {
		t.Errorf("unexpected dir output: %q", out)
	}
}

func TestOutput_ReturnsStdout(t *testing.T) {
	out, err := exec.Output("echo", []string{"hello"}, exec.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "hello" {
		t.Errorf("Output: got %q want %q", out, "hello")
	}
}

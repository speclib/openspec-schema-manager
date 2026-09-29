package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}

	bin := filepath.Join(t.TempDir(), "ossm")

	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building ossm: %v\n%s", err, out)
	}

	return bin
}

func runBinary(t *testing.T, bin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "TERM=dumb")

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Stdin = nil

	err := cmd.Run()
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("running %s %v: %v", bin, args, err)
	}

	return outBuf.String(), errBuf.String(), code
}

func TestVersionAnswersWithoutATerminal(t *testing.T) {
	bin := buildBinary(t)

	stdout, stderr, code := runBinary(t, bin, "--version")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Error("--version wrote nothing to standard output")
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Errorf("--version emitted an escape sequence, so it touched the terminal: %q", stdout)
	}
}

func TestHelpAnswersWithoutATerminal(t *testing.T) {
	bin := buildBinary(t)

	stdout, stderr, code := runBinary(t, bin, "--help")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr)
	}
	for _, want := range []string{"ossm [flags]", "-version", "-path"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("--help output does not mention %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Error("--help emitted an escape sequence, so it touched the terminal")
	}
}

func TestUnknownFlagIsRejected(t *testing.T) {
	bin := buildBinary(t)

	_, stderr, code := runBinary(t, bin, "--definitely-not-a-flag")

	if code == 0 {
		t.Error("exit code = 0, want non-zero for an unknown flag")
	}
	if stderr == "" {
		t.Error("an unknown flag produced no message on standard error")
	}
}

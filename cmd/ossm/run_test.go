package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{name: "version", args: []string{"--version"}, wantCode: 0, wantStdout: "0.1.0"},
		{name: "help", args: []string{"--help"}, wantCode: 0, wantStdout: "ossm [flags]"},
		{name: "unknown flag", args: []string{"--nope"}, wantCode: 2, wantStderr: "ossm:"},
		{name: "stray argument", args: []string{"leftover"}, wantCode: 2, wantStderr: "unexpected argument"},
		{name: "no arguments", args: nil, wantCode: 1, wantStderr: "not built yet"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			restore := version
			version = "0.1.0"
			defer func() { version = restore }()

			var stdout, stderr bytes.Buffer
			code := run(tc.args, &stdout, &stderr)

			if code != tc.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %s)", code, tc.wantCode, stderr.String())
			}
			if tc.wantStdout != "" && !strings.Contains(stdout.String(), tc.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), tc.wantStdout)
			}
			if tc.wantStderr != "" && !strings.Contains(stderr.String(), tc.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tc.wantStderr)
			}
		})
	}
}

func TestVersionIsTrimmed(t *testing.T) {
	restore := version
	version = "\n0.2.0\n"
	defer func() { version = restore }()

	var stdout, stderr bytes.Buffer
	if code := run([]string{"--version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	if got := stdout.String(); got != "0.2.0\n" {
		t.Errorf("stdout = %q, want %q", got, "0.2.0\n")
	}
}

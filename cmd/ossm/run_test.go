package main

import (
	"bytes"
	"os"
	"path/filepath"
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
		{name: "no terminal", args: nil, wantCode: 1, wantStderr: "not a terminal"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			restore := version
			version = "0.1.0"
			defer func() { version = restore }()

			var stdout, stderr bytes.Buffer
			code := run(tc.args, &stdout, &stderr, false)

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
	if code := run([]string{"--version"}, &stdout, &stderr, false); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	if got := stdout.String(); got != "0.2.0\n" {
		t.Errorf("stdout = %q, want %q", got, "0.2.0\n")
	}
}

func TestNoTerminalSaysSoWithoutEscapeSequences(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run(nil, &stdout, &stderr, false)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "not a terminal") {
		t.Errorf("stderr = %q, want it to name the problem", stderr.String())
	}
	if !strings.Contains(stderr.String(), "--version") {
		t.Errorf("stderr = %q, want it to point at the flags that need no terminal", stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), "\x1b[") {
		t.Error("a run with no terminal emitted an escape sequence")
	}
}

func TestFlagsAnswerEvenWithNoTerminal(t *testing.T) {
	for _, arg := range []string{"--version", "--help"} {
		var stdout, stderr bytes.Buffer

		if code := run([]string{arg}, &stdout, &stderr, false); code != 0 {
			t.Errorf("%s exit code = %d, want 0", arg, code)
		}
		if stdout.Len() == 0 {
			t.Errorf("%s wrote nothing to standard output", arg)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "not-a-terminal")
	if err != nil {
		t.Fatalf("creating a temporary file: %v", err)
	}
	defer file.Close()

	if isTerminal(file) {
		t.Error("a regular file was reported as a terminal")
	}

	closed, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("opening %s: %v", os.DevNull, err)
	}
	closed.Close()

	if isTerminal(closed) {
		t.Error("a closed file was reported as a terminal")
	}
}

func TestAppOptions(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "demo-app")
	if err := os.MkdirAll(filepath.Join(project, "openspec"), 0o755); err != nil {
		t.Fatalf("creating the project: %v", err)
	}
	outside := filepath.Join(root, "elsewhere")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("creating a directory outside any project: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(root, "state"))

	t.Run("inside a project", func(t *testing.T) {
		t.Chdir(project)

		got, err := appOptions(options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.InProject {
			t.Error("InProject is false inside a project")
		}
		if got.ProjectRoot == "" {
			t.Error("ProjectRoot is empty inside a project")
		}
		if got.Config.RegistryURL == "" {
			t.Error("the configuration was not loaded")
		}
		if got.Paths.ConfigFile == "" {
			t.Error("the paths were not resolved")
		}
	})

	t.Run("outside a project", func(t *testing.T) {
		t.Chdir(outside)

		got, err := appOptions(options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.InProject {
			t.Error("InProject is true outside a project")
		}
		if got.ProjectRoot != "" {
			t.Errorf("ProjectRoot = %q outside a project", got.ProjectRoot)
		}
	})

	t.Run("path flag points elsewhere", func(t *testing.T) {
		t.Chdir(outside)

		got, err := appOptions(options{path: project})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.InProject {
			t.Error("--path into a project did not detect it")
		}
		if got.WorkDir != project {
			t.Errorf("WorkDir = %q, want %q", got.WorkDir, project)
		}
	})
}

func TestAppOptionsReportsABadConfig(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "config", "ossm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte("nonsense_key: 1\n"), 0o644); err != nil {
		t.Fatalf("writing config.yml: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	if _, err := appOptions(options{}); err == nil {
		t.Fatal("expected an error for a configuration file that does not decode")
	}
}

func TestStartReportsABadConfig(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "config", "ossm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte("nonsense_key: 1\n"), 0o644); err != nil {
		t.Fatalf("writing config.yml: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	if err := start(options{}); err == nil {
		t.Fatal("expected start to report a configuration it cannot read")
	}
}

func TestRunReportsAFailureToStart(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "config", "ossm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte("nonsense_key: 1\n"), 0o644); err != nil {
		t.Fatalf("writing config.yml: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr, true); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "ossm:") {
		t.Errorf("stderr = %q, want it to name the tool and the problem", stderr.String())
	}
}

package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		wantVersion bool
		wantPath    string
		wantErr     error
		wantErrText string
	}{
		{name: "no arguments", args: nil},
		{name: "version", args: []string{"--version"}, wantVersion: true},
		{name: "version single dash", args: []string{"-version"}, wantVersion: true},
		{name: "path", args: []string{"--path", "/tmp/schemas/minimalist"}, wantPath: "/tmp/schemas/minimalist"},
		{name: "path and version", args: []string{"--path", "/x", "--version"}, wantVersion: true, wantPath: "/x"},
		{name: "help", args: []string{"--help"}, wantErr: errUsage},
		{name: "unknown flag", args: []string{"--nope"}, wantErrText: "flag provided but not defined"},
		{name: "stray argument", args: []string{"leftover"}, wantErrText: "unexpected argument"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			opts, err := parseFlags(tc.args, &out)

			switch {
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
				return
			case tc.wantErrText != "":
				if err == nil || !strings.Contains(err.Error(), tc.wantErrText) {
					t.Fatalf("got error %v, want one containing %q", err, tc.wantErrText)
				}
				return
			case err != nil:
				t.Fatalf("unexpected error: %v", err)
			}

			if opts.showVersion != tc.wantVersion {
				t.Errorf("showVersion = %v, want %v", opts.showVersion, tc.wantVersion)
			}
			if opts.path != tc.wantPath {
				t.Errorf("path = %q, want %q", opts.path, tc.wantPath)
			}
		})
	}
}

func TestHelpDescribesEveryFlag(t *testing.T) {
	var out bytes.Buffer
	if _, err := parseFlags([]string{"--help"}, &out); !errors.Is(err, errUsage) {
		t.Fatalf("got error %v, want errUsage", err)
	}

	help := out.String()
	for _, want := range []string{"ossm [flags]", "-version", "-path"} {
		if !strings.Contains(help, want) {
			t.Errorf("help output does not mention %q:\n%s", want, help)
		}
	}
}

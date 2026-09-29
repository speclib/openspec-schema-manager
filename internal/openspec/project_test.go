package openspec

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	project := filepath.Join(root, "demo-app")
	deep := filepath.Join(project, "src", "internal", "thing")
	if err := os.MkdirAll(filepath.Join(project, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("creating the project: %v", err)
	}
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("creating a subdirectory: %v", err)
	}

	outside := filepath.Join(root, "elsewhere")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("creating a directory outside any project: %v", err)
	}

	decoy := filepath.Join(root, "decoy")
	if err := os.MkdirAll(decoy, 0o755); err != nil {
		t.Fatalf("creating the decoy directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(decoy, "openspec"), []byte("a file, not a directory\n"), 0o644); err != nil {
		t.Fatalf("writing the decoy: %v", err)
	}

	cases := []struct {
		name     string
		dir      string
		wantRoot string
		wantErr  error
	}{
		{name: "at the project root", dir: project, wantRoot: project},
		{name: "deep inside the project", dir: deep, wantRoot: project},
		{name: "outside any project", dir: outside, wantErr: ErrNoProject},
		{name: "openspec is a file", dir: decoy, wantErr: ErrNoProject},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := FindProjectRoot(tc.dir)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got (%q, %v), want error %v", got, err, tc.wantErr)
				}
				if InProject(tc.dir) {
					t.Error("InProject says yes where FindProjectRoot says no")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantRoot {
				t.Errorf("root = %q, want %q", got, tc.wantRoot)
			}
			if !InProject(tc.dir) {
				t.Error("InProject says no where FindProjectRoot says yes")
			}
		})
	}
}

func TestFindProjectRootStopsAtTheFilesystemRoot(t *testing.T) {
	t.Parallel()

	if _, err := FindProjectRoot(string(filepath.Separator)); err != nil && !errors.Is(err, ErrNoProject) {
		t.Fatalf("walking up from the filesystem root gave %v, want nil or ErrNoProject", err)
	}
}

func TestFindProjectRootResolvesARelativePath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "openspec"), 0o755); err != nil {
		t.Fatalf("creating the project: %v", err)
	}

	t.Chdir(root)

	got, err := FindProjectRoot(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("resolving %s: %v", root, err)
	}
	gotResolved, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("resolving %s: %v", got, err)
	}
	if gotResolved != want {
		t.Errorf("root = %q, want %q", gotResolved, want)
	}
}

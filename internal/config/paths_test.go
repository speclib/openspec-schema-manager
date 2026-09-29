package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeEnv(home string, vars map[string]string) env {
	return env{
		getenv: func(k string) string { return vars[k] },
		home:   func() (string, error) { return home, nil },
	}
}

func TestPaths(t *testing.T) {
	t.Parallel()

	const home = "/home/tester"

	cases := []struct {
		name string
		vars map[string]string
		want Paths
	}{
		{
			name: "no variables set",
			vars: nil,
			want: Paths{
				ConfigFile:   filepath.Join(home, ".config", "ossm", "config.yml"),
				RegistryFile: filepath.Join(home, ".cache", "ossm", "registry.json"),
				SchemaCache:  filepath.Join(home, ".cache", "ossm", "schemas"),
				RecentsFile:  filepath.Join(home, ".local", "state", "ossm", "recents.json"),
				DraftsDir:    filepath.Join(home, ".local", "state", "ossm", "drafts"),
			},
		},
		{
			name: "cache root relocated",
			vars: map[string]string{"XDG_CACHE_HOME": "/var/cache/me"},
			want: Paths{
				ConfigFile:   filepath.Join(home, ".config", "ossm", "config.yml"),
				RegistryFile: filepath.Join("/var/cache/me", "ossm", "registry.json"),
				SchemaCache:  filepath.Join("/var/cache/me", "ossm", "schemas"),
				RecentsFile:  filepath.Join(home, ".local", "state", "ossm", "recents.json"),
				DraftsDir:    filepath.Join(home, ".local", "state", "ossm", "drafts"),
			},
		},
		{
			name: "state root relocated",
			vars: map[string]string{"XDG_STATE_HOME": "/var/state/me"},
			want: Paths{
				ConfigFile:   filepath.Join(home, ".config", "ossm", "config.yml"),
				RegistryFile: filepath.Join(home, ".cache", "ossm", "registry.json"),
				SchemaCache:  filepath.Join(home, ".cache", "ossm", "schemas"),
				RecentsFile:  filepath.Join("/var/state/me", "ossm", "recents.json"),
				DraftsDir:    filepath.Join("/var/state/me", "ossm", "drafts"),
			},
		},
		{
			name: "config root relocated",
			vars: map[string]string{"XDG_CONFIG_HOME": "/etc/xdg/me"},
			want: Paths{
				ConfigFile:   filepath.Join("/etc/xdg/me", "ossm", "config.yml"),
				RegistryFile: filepath.Join(home, ".cache", "ossm", "registry.json"),
				SchemaCache:  filepath.Join(home, ".cache", "ossm", "schemas"),
				RecentsFile:  filepath.Join(home, ".local", "state", "ossm", "recents.json"),
				DraftsDir:    filepath.Join(home, ".local", "state", "ossm", "drafts"),
			},
		},
		{
			name: "an empty variable falls back",
			vars: map[string]string{"XDG_CACHE_HOME": ""},
			want: Paths{
				ConfigFile:   filepath.Join(home, ".config", "ossm", "config.yml"),
				RegistryFile: filepath.Join(home, ".cache", "ossm", "registry.json"),
				SchemaCache:  filepath.Join(home, ".cache", "ossm", "schemas"),
				RecentsFile:  filepath.Join(home, ".local", "state", "ossm", "recents.json"),
				DraftsDir:    filepath.Join(home, ".local", "state", "ossm", "drafts"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := fakeEnv(home, tc.vars).paths()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("paths mismatch\n got: %+v\nwant: %+v", got, tc.want)
			}
		})
	}
}

func TestResolvePathsUsesTheProcessEnvironment(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/ossm-config-root")

	got, err := ResolvePaths()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join("/tmp/ossm-config-root", "ossm", "config.yml")
	if got.ConfigFile != want {
		t.Errorf("ConfigFile = %q, want %q", got.ConfigFile, want)
	}
}

func TestPathsReportsAHomeItCannotFind(t *testing.T) {
	t.Parallel()

	broken := env{
		getenv: func(string) string { return "" },
		home:   func() (string, error) { return "", os.ErrNotExist },
	}

	if _, err := broken.paths(); err == nil {
		t.Fatal("expected an error when the home directory cannot be resolved")
	}
}

func TestEnsureDirCreatesAndIsIdempotent(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("second call: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("EnsureDir did not create a directory")
	}
}

func TestEnsureParentCreatesTheParentOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "state", "ossm", "recents.json")

	if err := EnsureParent(target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(target)); err != nil {
		t.Errorf("parent directory was not created: %v", err)
	}
	if _, err := os.Stat(target); err == nil {
		t.Error("EnsureParent created the file itself")
	}
}

func TestLoadCreatesNothing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	e := fakeEnv(root, map[string]string{
		"XDG_CONFIG_HOME": filepath.Join(root, "config"),
		"XDG_CACHE_HOME":  filepath.Join(root, "cache"),
		"XDG_STATE_HOME":  filepath.Join(root, "state"),
	})

	if _, _, err := e.load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading %s: %v", root, err)
	}
	if len(entries) != 0 {
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Errorf("loading created %s; it must create nothing", strings.Join(names, ", "))
	}
}

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeConfig(t *testing.T, home, body string) env {
	t.Helper()

	root := t.TempDir()
	dir := filepath.Join(root, "ossm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(body), 0o644); err != nil {
		t.Fatalf("writing config.yml: %v", err)
	}

	return fakeEnv(home, map[string]string{"XDG_CONFIG_HOME": root})
}

func TestDefaultsWithNoFile(t *testing.T) {
	t.Parallel()

	e := fakeEnv("/home/tester", map[string]string{"XDG_CONFIG_HOME": t.TempDir()})

	cfg, _, err := e.load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Defaults()
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("config mismatch\n got: %+v\nwant: %+v", cfg, want)
	}
	if cfg.RegistryURL != DefaultRegistryURL {
		t.Errorf("RegistryURL = %q, want the published registry address", cfg.RegistryURL)
	}
	if cfg.RegistryTTL != 24*time.Hour {
		t.Errorf("RegistryTTL = %v, want 24h", cfg.RegistryTTL)
	}
	if cfg.RecentsCap != 20 {
		t.Errorf("RecentsCap = %d, want 20", cfg.RecentsCap)
	}
	if len(cfg.SchemasDirs) != 0 {
		t.Errorf("SchemasDirs = %v, want empty", cfg.SchemasDirs)
	}
}

func TestPartialFileLeavesTheRestAtDefaults(t *testing.T) {
	t.Parallel()

	e := writeConfig(t, "/home/tester", "registry_ttl: 30m\n")

	cfg, _, err := e.load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RegistryTTL != 30*time.Minute {
		t.Errorf("RegistryTTL = %v, want 30m", cfg.RegistryTTL)
	}
	if cfg.RegistryURL != DefaultRegistryURL {
		t.Errorf("RegistryURL = %q, want the default", cfg.RegistryURL)
	}
	if cfg.RecentsCap != DefaultRecentsCap {
		t.Errorf("RecentsCap = %d, want the default", cfg.RecentsCap)
	}
}

func TestEveryKeyTakesEffect(t *testing.T) {
	t.Parallel()

	e := writeConfig(t, "/home/tester", strings.Join([]string{
		"registry_url: file:///srv/mirror/openspec-schemas.json",
		"registry_ttl: 0s",
		"recents_cap: 5",
		"schemas_dirs:",
		"  - /srv/schemas",
		"  - ~/work/schemas",
	}, "\n")+"\n")

	cfg, _, err := e.load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RegistryURL != "file:///srv/mirror/openspec-schemas.json" {
		t.Errorf("RegistryURL = %q", cfg.RegistryURL)
	}
	if cfg.RegistryTTL != 0 {
		t.Errorf("RegistryTTL = %v, want 0", cfg.RegistryTTL)
	}
	if cfg.RecentsCap != 5 {
		t.Errorf("RecentsCap = %d, want 5", cfg.RecentsCap)
	}

	want := []string{"/srv/schemas", filepath.Join("/home/tester", "work", "schemas")}
	if !reflect.DeepEqual(cfg.SchemasDirs, want) {
		t.Errorf("SchemasDirs = %v, want %v", cfg.SchemasDirs, want)
	}
}

func TestTildeExpansion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "leading tilde slash", in: "~/work/schemas", want: filepath.Join("/home/tester", "work", "schemas")},
		{name: "bare tilde", in: "~", want: "/home/tester"},
		{name: "tilde in the middle", in: "/srv/~backup/schemas", want: "/srv/~backup/schemas"},
		{name: "tilde in a name", in: "~backup/schemas", want: "~backup/schemas"},
		{name: "already absolute", in: "/srv/schemas", want: "/srv/schemas"},
		{name: "relative", in: "schemas", want: "schemas"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := fakeEnv("/home/tester", nil).expandTilde(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("expandTilde(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRejectsAFileItCannotTrust(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "not valid yaml", body: "registry_ttl: [unclosed\n", wantErr: "config.yml"},
		{name: "top level is a list", body: "- one\n- two\n", wantErr: "config.yml"},
		{name: "top level is a string", body: "just a string\n", wantErr: "config.yml"},
		{name: "unknown key", body: "registry_urls: https://example.test/x\n", wantErr: "registry_urls"},
		{name: "ttl is a bare number", body: "registry_ttl: 30\n", wantErr: "registry_ttl"},
		{name: "ttl is nonsense", body: "registry_ttl: soonish\n", wantErr: "registry_ttl"},
		{name: "ttl is negative", body: "registry_ttl: -1h\n", wantErr: "registry_ttl"},
		{name: "negative recents cap", body: "recents_cap: -1\n", wantErr: "recents_cap"},
		{name: "recents cap is a string", body: "recents_cap: many\n", wantErr: "config.yml"},
		{name: "schemas dirs is a string", body: "schemas_dirs: /srv/schemas\n", wantErr: "config.yml"},
		{name: "empty registry url", body: "registry_url: \"\"\n", wantErr: "registry_url"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := writeConfig(t, "/home/tester", tc.body).load()
			if err == nil {
				t.Fatalf("expected an error for %q", tc.body)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not mention %q", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), "config.yml") {
				t.Errorf("error %q does not name the file", err)
			}
		})
	}
}

func TestAnEmptyFileIsValid(t *testing.T) {
	t.Parallel()

	cfg, _, err := writeConfig(t, "/home/tester", "").load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(cfg, Defaults()) {
		t.Errorf("an empty file should leave the defaults, got %+v", cfg)
	}
}

func TestAnUnreadableFileIsReported(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unreadable file is still readable")
	}

	root := t.TempDir()
	dir := filepath.Join(root, "ossm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating %s: %v", dir, err)
	}
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("registry_ttl: 1h\n"), 0o000); err != nil {
		t.Fatalf("writing config.yml: %v", err)
	}

	_, _, err := fakeEnv("/home/tester", map[string]string{"XDG_CONFIG_HOME": root}).load()
	if err == nil {
		t.Fatal("expected an error for an unreadable file")
	}
	if !strings.Contains(err.Error(), "config.yml") {
		t.Errorf("error %q does not name the file", err)
	}
}

func TestLoadUsesTheProcessEnvironment(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	cfg, paths, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RegistryURL != DefaultRegistryURL {
		t.Errorf("RegistryURL = %q, want the default", cfg.RegistryURL)
	}
	if paths.ConfigFile != filepath.Join(root, "ossm", "config.yml") {
		t.Errorf("ConfigFile = %q", paths.ConfigFile)
	}
}

func TestTheExampleConfigDecodes(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(filepath.Join("..", "..", "docs", "config.example.yml"))
	if err != nil {
		t.Fatalf("reading the example: %v", err)
	}

	cfg, _, err := writeConfig(t, "/home/tester", string(body)).load()
	if err != nil {
		t.Fatalf("the example config does not decode: %v", err)
	}

	for _, key := range []string{"registry_url", "registry_ttl", "schemas_dirs", "recents_cap"} {
		if !strings.Contains(string(body), key+":") {
			t.Errorf("the example config does not mention %q", key)
		}
	}

	if cfg.RegistryURL == "" {
		t.Error("the example config left the registry URL empty")
	}
}

func brokenHomeEnv(vars map[string]string) env {
	return env{
		getenv: func(k string) string { return vars[k] },
		home:   func() (string, error) { return "", os.ErrNotExist },
	}
}

func TestATildeWithNoHomeIsReported(t *testing.T) {
	t.Parallel()

	if _, err := brokenHomeEnv(nil).expandTilde("~/work/schemas"); err == nil {
		t.Fatal("expected an error when the home directory cannot be resolved")
	}
	if _, err := brokenHomeEnv(nil).expandTilde("~"); err == nil {
		t.Fatal("expected an error for a bare tilde with no home")
	}
	if got, err := brokenHomeEnv(nil).expandTilde("/srv/schemas"); err != nil || got != "/srv/schemas" {
		t.Errorf("an absolute path must not need a home: got %q, %v", got, err)
	}
}

func TestATildeInTheConfigWithNoHomeIsReported(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := filepath.Join(root, "ossm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte("schemas_dirs:\n  - ~/work\n"), 0o644); err != nil {
		t.Fatalf("writing config.yml: %v", err)
	}

	e := brokenHomeEnv(map[string]string{
		"XDG_CONFIG_HOME": root,
		"XDG_CACHE_HOME":  root,
		"XDG_STATE_HOME":  root,
	})

	_, _, err := e.load()
	if err == nil {
		t.Fatal("expected an error when a tilde cannot be expanded")
	}
	if !strings.Contains(err.Error(), "schemas_dirs") {
		t.Errorf("error %q does not name the field", err)
	}
}

func TestLoadReportsAHomeItCannotFind(t *testing.T) {
	t.Parallel()

	if _, _, err := brokenHomeEnv(nil).load(); err == nil {
		t.Fatal("expected an error when no root can be resolved")
	}
}

func TestPathsReportsAMissingHomePerRoot(t *testing.T) {
	t.Parallel()

	cases := []map[string]string{
		{},
		{"XDG_CONFIG_HOME": "/etc/xdg"},
		{"XDG_CONFIG_HOME": "/etc/xdg", "XDG_CACHE_HOME": "/var/cache"},
	}

	for i, vars := range cases {
		if _, err := brokenHomeEnv(vars).paths(); err == nil {
			t.Errorf("case %d: expected an error when a root falls back to a home that cannot be resolved", i)
		}
	}
}

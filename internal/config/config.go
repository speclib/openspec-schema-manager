package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultRegistryURL = "https://registry.speclib.org/api/v1/openspec-schemas.json"
	DefaultRegistryTTL = 24 * time.Hour
	DefaultRecentsCap  = 20
)

type Config struct {
	RegistryURL string
	RegistryTTL time.Duration
	SchemasDirs []string
	RecentsCap  int
}

func Defaults() Config {
	return Config{
		RegistryURL: DefaultRegistryURL,
		RegistryTTL: DefaultRegistryTTL,
		SchemasDirs: []string{},
		RecentsCap:  DefaultRecentsCap,
	}
}

type file struct {
	RegistryURL *string   `yaml:"registry_url"`
	RegistryTTL *string   `yaml:"registry_ttl"`
	SchemasDirs *[]string `yaml:"schemas_dirs"`
	RecentsCap  *int      `yaml:"recents_cap"`
}

func Load() (Config, Paths, error) {
	return osEnv().load()
}

func (e env) load() (Config, Paths, error) {
	paths, err := e.paths()
	if err != nil {
		return Config{}, Paths{}, err
	}

	cfg, err := e.loadFile(paths.ConfigFile)
	if err != nil {
		return Config{}, paths, err
	}

	return cfg, paths, nil
}

func (e env) loadFile(path string) (Config, error) {
	cfg := Defaults()

	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return cfg, nil
	case err != nil:
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var parsed file
	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true)

	if err := dec.Decode(&parsed); err != nil && err.Error() != "EOF" {
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}

	if parsed.RegistryURL != nil {
		if strings.TrimSpace(*parsed.RegistryURL) == "" {
			return Config{}, fmt.Errorf("reading %s: registry_url is empty; remove the key to use the default", path)
		}
		cfg.RegistryURL = strings.TrimSpace(*parsed.RegistryURL)
	}

	if parsed.RegistryTTL != nil {
		ttl, err := time.ParseDuration(*parsed.RegistryTTL)
		if err != nil {
			return Config{}, fmt.Errorf("reading %s: registry_ttl %q is not a duration; write it as 24h, 30m or 0s", path, *parsed.RegistryTTL)
		}
		if ttl < 0 {
			return Config{}, fmt.Errorf("reading %s: registry_ttl %q is negative", path, *parsed.RegistryTTL)
		}
		cfg.RegistryTTL = ttl
	}

	if parsed.RecentsCap != nil {
		if *parsed.RecentsCap < 0 {
			return Config{}, fmt.Errorf("reading %s: recents_cap is %d; a cap below zero has no meaning", path, *parsed.RecentsCap)
		}
		cfg.RecentsCap = *parsed.RecentsCap
	}

	if parsed.SchemasDirs != nil {
		dirs := make([]string, 0, len(*parsed.SchemasDirs))
		for _, dir := range *parsed.SchemasDirs {
			expanded, err := e.expandTilde(dir)
			if err != nil {
				return Config{}, fmt.Errorf("reading %s: schemas_dirs entry %q: %w", path, dir, err)
			}
			dirs = append(dirs, expanded)
		}
		cfg.SchemasDirs = dirs
	}

	return cfg, nil
}

func (e env) expandTilde(dir string) (string, error) {
	if dir != "~" && !strings.HasPrefix(dir, "~/") && !strings.HasPrefix(dir, `~\`) {
		return dir, nil
	}

	home, err := e.home()
	if err != nil {
		return "", err
	}

	if dir == "~" {
		return home, nil
	}

	return filepath.Join(home, filepath.FromSlash(dir[2:])), nil
}

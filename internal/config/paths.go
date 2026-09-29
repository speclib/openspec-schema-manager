package config

import (
	"os"
	"path/filepath"
)

type env struct {
	getenv func(string) string
	home   func() (string, error)
}

func osEnv() env {
	return env{getenv: os.Getenv, home: os.UserHomeDir}
}

type Paths struct {
	ConfigFile   string
	RegistryFile string
	SchemaCache  string
	RecentsFile  string
	DraftsDir    string
}

func (e env) root(variable, fallback string) (string, error) {
	if v := e.getenv(variable); v != "" {
		return v, nil
	}

	home, err := e.home()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, filepath.FromSlash(fallback)), nil
}

func (e env) paths() (Paths, error) {
	configRoot, err := e.root("XDG_CONFIG_HOME", ".config")
	if err != nil {
		return Paths{}, err
	}

	cacheRoot, err := e.root("XDG_CACHE_HOME", ".cache")
	if err != nil {
		return Paths{}, err
	}

	stateRoot, err := e.root("XDG_STATE_HOME", ".local/state")
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ConfigFile:   filepath.Join(configRoot, "ossm", "config.yml"),
		RegistryFile: filepath.Join(cacheRoot, "ossm", "registry.json"),
		SchemaCache:  filepath.Join(cacheRoot, "ossm", "schemas"),
		RecentsFile:  filepath.Join(stateRoot, "ossm", "recents.json"),
		DraftsDir:    filepath.Join(stateRoot, "ossm", "drafts"),
	}, nil
}

func ResolvePaths() (Paths, error) {
	return osEnv().paths()
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func EnsureParent(file string) error {
	return EnsureDir(filepath.Dir(file))
}

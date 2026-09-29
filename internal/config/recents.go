package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Recents struct {
	Path string
	Cap  int
}

type recentsFile struct {
	Paths []string `json:"paths"`
}

func (r Recents) Read() []string {
	raw, err := os.ReadFile(r.Path)
	if err != nil {
		return nil
	}

	var file recentsFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil
	}

	return r.trim(file.Paths)
}

func (r Recents) Add(path string) ([]string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		absolute = path
	}

	paths := []string{absolute}
	for _, existing := range r.Read() {
		if existing != absolute {
			paths = append(paths, existing)
		}
	}

	paths = r.trim(paths)

	return paths, r.write(paths)
}

func (r Recents) trim(paths []string) []string {
	cap := r.Cap
	if cap < 0 {
		cap = 0
	}

	if len(paths) > cap {
		paths = paths[:cap]
	}

	return paths
}

func (r Recents) write(paths []string) error {
	if paths == nil {
		paths = []string{}
	}

	raw, err := json.Marshal(recentsFile{Paths: paths})
	if err != nil {
		return err
	}

	if err := EnsureParent(r.Path); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(r.Path), ".recents-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), r.Path)
}

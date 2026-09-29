package openspec

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNoProject = errors.New("not inside an OpenSpec project")

func FindProjectRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for {
		info, err := os.Stat(filepath.Join(abs, "openspec"))
		if err == nil && info.IsDir() {
			return abs, nil
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			return "", ErrNoProject
		}
		abs = parent
	}
}

func InProject(dir string) bool {
	_, err := FindProjectRoot(dir)
	return err == nil
}

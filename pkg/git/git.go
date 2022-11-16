package git

import (
	"os"
	"path/filepath"

	"github.com/fioncat/gitzombie/pkg/errors"
)

var ErrNotGit = errors.New("you are not int a git repository")

func EnsureCurrent() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, EnsurePath(dir)
}

func EnsurePath(dir string) error {
	gitDir := filepath.Join(dir, ".git")
	stat, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotGit
		}
		return errors.Trace(err, "check git exists")
	}
	if !stat.IsDir() {
		return ErrNotGit
	}
	return nil
}

package git

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
)

type DiscoverDir struct {
	Name string
	Path string
}

func Discover(rootDir string) ([]*DiscoverDir, error) {
	if !filepath.IsAbs(rootDir) {
		return nil, fmt.Errorf("require aps path, found %q", rootDir)
	}
	var dirs []*DiscoverDir
	err := filepath.WalkDir(rootDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		gitDir := filepath.Join(path, ".git")
		exists, err := osutil.DirExists(gitDir)
		if err != nil {
			return errors.Trace(err, "check exists for %q", gitDir)
		}
		if !exists {
			return nil
		}

		name, err := filepath.Rel(rootDir, path)
		if err != nil {
			return errors.Trace(err, "convert rel path")
		}
		dirs = append(dirs, &DiscoverDir{
			Name: name,
			Path: path,
		})
		return nil
	})
	return dirs, err
}

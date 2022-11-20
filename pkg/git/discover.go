package git

import (
	"fmt"
	"os"
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
		return nil, fmt.Errorf("require abs path, found %q", rootDir)
	}
	stack := []string{rootDir}
	var dirs []*DiscoverDir
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		es, err := os.ReadDir(cur)
		if err != nil {
			return nil, err
		}
		for _, e := range es {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(cur, e.Name())
			gitDir := filepath.Join(dir, ".git")
			exists, err := osutil.DirExists(gitDir)
			if err != nil {
				return nil, errors.Trace(err, "check git dir")
			}
			if exists {
				name, err := filepath.Rel(rootDir, dir)
				if err != nil {
					return nil, errors.Trace(err, "convert rel path")
				}
				dirs = append(dirs, &DiscoverDir{
					Name: name,
					Path: dir,
				})
				continue
			}
			stack = append(stack, dir)
		}
	}
	return dirs, nil
}

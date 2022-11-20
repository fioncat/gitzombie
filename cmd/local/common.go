package local

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

func Select(rootDir, name string) (*core.Repository, error) {
	if strings.HasSuffix(name, "/") {
		searchDir := filepath.Join(rootDir, name)
		foundDirs, err := git.Discover(searchDir)
		if err != nil {
			return nil, err
		}
		items := make([]string, len(foundDirs))
		for i, foundDir := range foundDirs {
			items[i] = foundDir.Name
		}
		idx, err := term.FuzzySearch("repo", items)
		if err != nil {
			return nil, err
		}
		item := items[idx]
		name = filepath.Join(name, item)
		return core.NewLocalRepository(rootDir, name)
	}
	name = strings.Trim(name, "/")
	if name != "" {
		gitPath := filepath.Join(rootDir, name, ".git")
		exists, err := osutil.DirExists(gitPath)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("cannot find %s", name)
		}
		return core.NewLocalRepository(rootDir, name)
	}
	repos, err := core.DiscoverLocalRepositories(rootDir)
	if err != nil {
		return nil, err
	}
	items := make([]string, len(repos))
	for i, repo := range repos {
		items[i] = repo.Name
	}
	idx, err := term.FuzzySearch("repo", items)
	if err != nil {
		return nil, err
	}
	return repos[idx], nil
}

package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

type LocalRepository struct {
	Name string

	Group string
	Base  string

	Path      string
	GroupPath string
}

func (repo *LocalRepository) Setenv() error {
	return osutil.Setenv(map[string]string{
		"REPO_NAME":       repo.Name,
		"REPO_BASE":       repo.Base,
		"REPO_GROUP":      repo.Group,
		"REPO_PATH":       repo.Path,
		"REPO_GROUP_PATH": repo.GroupPath,
	})
}

func NewLocalRepository(rootDir, name string) (*LocalRepository, error) {
	name = strings.Trim(name, "/")
	group, base := SplitGroup(name)
	if group == "" || name == "" {
		return nil, fmt.Errorf("invalid repo name %q, must have a group", name)
	}

	return &LocalRepository{
		Name:      name,
		Group:     group,
		Base:      base,
		Path:      filepath.Join(rootDir, name),
		GroupPath: filepath.Join(rootDir, group),
	}, nil

}

func DiscoverLocalRepositories(rootDir string) ([]*LocalRepository, error) {
	dirs, err := git.Discover(rootDir)
	if err != nil {
		return nil, err
	}
	repos := make([]*LocalRepository, len(dirs))
	for i, dir := range dirs {
		repo, err := NewLocalRepository(rootDir, dir.Name)
		if err != nil {
			return nil, err
		}
		repos[i] = repo
	}
	return repos, nil
}

func SelectLocalRepository(rootDir, name string) (*LocalRepository, error) {
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
		return NewLocalRepository(rootDir, name)
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
		return NewLocalRepository(rootDir, name)
	}
	repos, err := DiscoverLocalRepositories(rootDir)
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

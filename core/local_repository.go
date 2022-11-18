package core

import "github.com/fioncat/gitzombie/pkg/osutil"

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

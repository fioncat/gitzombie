package play

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
)

func list() ([]*core.LocalRepository, error) {
	rootDir := config.Get().Playground
	return core.DiscoverLocalRepositories(rootDir)
}

func compRepo(_ []string) (*app.CompResult, error) {
	repos, err := list()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.Name
	}
	return &app.CompResult{Items: names}, nil
}

func compGroup(_ []string) (*app.CompResult, error) {
	repos, err := list()
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(repos))
	var groups []string
	for _, repo := range repos {
		if _, ok := set[repo.Group]; ok {
			continue
		}
		set[repo.Group] = struct{}{}
		groups = append(groups, repo.Group+"/")
	}
	return &app.CompResult{
		Items: groups,
		Flag:  app.CompNoSpaceFlag,
	}, nil
}

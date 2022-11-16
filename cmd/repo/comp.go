package repo

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/core"
)

func compRemote(_ common.Args) (*common.CompResult, error) {
	remotes, err := core.ListRemoteNames()
	return &common.CompResult{Items: remotes}, err
}

func listRepos(args common.Args) ([]*core.Repository, error) {
	remote := args.Get(0)
	if remote == "" {
		return nil, nil
	}
	store, err := core.NewRepositoryStorage()
	if err != nil {
		return nil, err
	}
	return store.List(remote), nil
}

func compRepo(args common.Args) (*common.CompResult, error) {
	repos, err := listRepos(args)
	if err != nil {
		return nil, err
	}

	items := make([]string, len(repos))
	for i, repo := range repos {
		items[i] = repo.Name
	}

	return &common.CompResult{Items: items}, nil
}

func compGroup(args common.Args) (*common.CompResult, error) {
	repos, err := listRepos(args)
	if err != nil {
		return nil, err
	}

	return &common.CompResult{
		Items: convertToGroups(repos),
		Flag:  common.CompNoSpaceFlag,
	}, nil
}

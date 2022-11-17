package gitops

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
)

func CompRemote(_ []string) (*app.CompResult, error) {
	_, err := git.EnsureCurrent()
	if err != nil {
		return app.EmptyCompResult, nil
	}
	remotes, err := git.ListRemotes(git.Mute)
	if err != nil {
		return nil, err
	}
	return &app.CompResult{Items: remotes}, nil
}

func CompLocalBranch(_ []string) (*app.CompResult, error) {
	branches, err := git.ListLocalBranches(git.Mute)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(branches))
	for i, branch := range branches {
		names[i] = branch.Name
	}
	return &app.CompResult{Items: names}, nil
}

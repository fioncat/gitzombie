package gitops

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/pkg/git"
)

func CompRemote(_ common.Args) (*common.CompResult, error) {
	_, err := git.EnsureCurrent()
	if err != nil {
		return common.EmptyCompResult, nil
	}
	remotes, err := git.ListRemotes(git.Mute)
	if err != nil {
		return nil, err
	}
	return &common.CompResult{Items: remotes}, nil
}

func CompLocalBranch(_ common.Args) (*common.CompResult, error) {
	branches, err := git.ListLocalBranches(git.Mute)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(branches))
	for i, branch := range branches {
		names[i] = branch.Name
	}
	return &common.CompResult{Items: names}, nil
}

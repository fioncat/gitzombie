package git

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/pkg/git"
)

func compRemote(_ common.Args) (*common.CompResult, error) {
	_, err := git.EnsureCurrent()
	if err != nil {
		return common.EmptyCompResult, nil
	}
	remotes, err := git.ListRemotes(git.Mute)
	return &common.CompResult{Items: remotes}, nil
}

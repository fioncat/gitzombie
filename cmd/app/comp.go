package app

import (
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

const CompNoSpaceFlag = cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp

type CompResult struct {
	Items []string
	Flag  cobra.ShellCompDirective
}

var EmptyCompResult = &CompResult{
	Flag: cobra.ShellCompDirectiveNoFileComp,
}

type CompAction func(args []string) (*CompResult, error)

func Comp(actions ...CompAction) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(actions) == 0 {
		return nil
	}
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		err := config.Init()
		if err != nil {
			term.Warn("complete: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}
		idx := len(args)
		if idx >= len(actions) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		action := actions[idx]
		result, err := action(args)
		if err != nil {
			term.Warn("complete: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}
		if result.Flag <= 0 {
			result.Flag = cobra.ShellCompDirectiveNoFileComp
		}
		return result.Items, result.Flag
	}
}
func CompGitRemote(_ []string) (*CompResult, error) {
	_, err := git.EnsureCurrent()
	if err != nil {
		return EmptyCompResult, nil
	}
	remotes, err := git.ListRemotes(git.Mute)
	if err != nil {
		return nil, err
	}
	return &CompResult{Items: remotes}, nil
}

func CompGitLocalBranch(_ []string) (*CompResult, error) {
	branches, err := git.ListLocalBranches(git.Mute)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(branches))
	for i, branch := range branches {
		names[i] = branch.Name
	}
	return &CompResult{Items: names}, nil
}

func compListRepos(args []string) ([]*core.Repository, error) {
	remote := args[0]
	if remote == "" {
		return nil, nil
	}
	store, err := core.NewRepositoryStorage()
	if err != nil {
		return nil, err
	}
	return store.List(remote), nil
}

func CompRemote(_ []string) (*CompResult, error) {
	remotes, err := core.ListRemoteNames()
	return &CompResult{Items: remotes}, err
}

func CompRepo(args []string) (*CompResult, error) {
	repos, err := compListRepos(args)
	if err != nil {
		return nil, err
	}

	items := make([]string, len(repos))
	for i, repo := range repos {
		items[i] = repo.Name
	}

	return &CompResult{Items: items}, nil
}

func CompGroup(args []string) (*CompResult, error) {
	repos, err := compListRepos(args)
	if err != nil {
		return nil, err
	}

	return &CompResult{
		Items: core.ConvertToGroups(repos),
		Flag:  CompNoSpaceFlag,
	}, nil
}

package tools

import (
	"fmt"

	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
)

func getTarget(branch, remote string, upstream bool) (string, string, error) {
	if upstream {
		if remote == "" {
			remote = "upstream"
		}
		err := ensureUpstream(remote)
		if err != nil {
			return "", "", err
		}
	} else {
		if remote == "" {
			remote = "origin"
		}
	}

	err := git.Fetch(remote, true, false, git.Default)
	if err != nil {
		return "", "", err
	}
	if branch == "" {
		mainBranch, err := git.GetDefaultBranch(remote, git.Default)
		if err != nil {
			return "", "", err
		}
		term.Println("use default branch ", term.Style(mainBranch, "green"))
		branch = mainBranch
	}

	return fmt.Sprintf("%s/%s", remote, branch), branch, nil
}

func ensureUpstream(upstreamGitRemote string) error {
	gitRemotes, err := git.ListRemotes(git.Default)
	if err != nil {
		return err
	}
	var found bool
	for _, gitRemote := range gitRemotes {
		if gitRemote == upstreamGitRemote {
			found = true
			break
		}
	}
	if found {
		return nil
	}
	term.PrintOperation("cannot find remote upstream, try to fetch")
	store, err := core.NewRepositoryStorage()
	if err != nil {
		return errors.Trace(err, "init repo store")
	}
	store.ReadOnly()

	repo, err := store.GetCurrent()
	if err != nil {
		return errors.Trace(err, "get current repo")
	}

	remote, err := core.GetRemote(repo.Remote)
	if err != nil {
		return err
	}

	var apiRepo *api.Repository
	err = api.Exec("fetch upstream", remote, func(p api.Provider) error {
		apiRepo, err = p.GetRepository(repo.Name)
		return err
	})
	if err != nil {
		return err
	}
	if apiRepo.Upstream == nil {
		return fmt.Errorf("repo %q does not have an upstream", repo.FullName())
	}

	upRepo, err := core.WorkspaceRepository(remote, apiRepo.Upstream.Name)
	if err != nil {
		return err
	}

	url, err := remote.GetCloneURL(upRepo)
	if err != nil {
		return err
	}

	term.ConfirmExit("Do you want to set remote %s to %q", upstreamGitRemote, url)
	return git.Exec([]string{"remote", "add", "upstream", url}, git.Default)
}

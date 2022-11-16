package repo

import (
	"fmt"
	"os"

	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Home struct {
	search bool
}

func (home *Home) Use() string    { return "repo remote [repo]" }
func (home *Home) Desc() string   { return "Print the home path of a repo" }
func (home *Home) Action() string { return "home" }

func (home *Home) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.RangeArgs(1, 2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compRepo)
	cmd.Flags().BoolVarP(&home.search, "search", "s", false, "search from remote")
}

func (home *Home) Run(ctx *Context, args common.Args) error {
	var repo *core.Repository
	var err error
	if home.search {
		repo, err = home.searchRepo(ctx, args)
	} else {
		repo, err = getLocal(ctx, args.Get(1))
	}
	if err != nil {
		return err
	}

	err = home.ensureRepo(ctx, ctx.remote, repo)
	if err != nil {
		return err
	}
	repo.View++
	fmt.Println(repo.Path)
	return nil
}

func (home *Home) searchRepo(ctx *Context, args common.Args) (*core.Repository, error) {
	apiRepo, err := apiSearch(ctx, args.Get(1))
	if err != nil {
		return nil, err
	}

	repo := ctx.store.GetByName(ctx.remote.Name, apiRepo.Name)
	if repo == nil {
		repo, err = core.CreateRepository(ctx.remote, apiRepo.Name)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}

func (home *Home) ensureRepo(ctx *Context, remote *core.Remote, repo *core.Repository) error {
	url, err := remote.GetCloneURL(repo)
	if err != nil {
		return errors.Trace(err, "get clone url")
	}
	stat, err := os.Stat(repo.Path)
	switch {
	case os.IsNotExist(err):
		term.ConfirmExit("repo %s does not exists, do you want to clone it", repo.FullName())
		err = git.Clone(url, repo.Path, git.Default)
		if err != nil {
			return err
		}
		user, email := remote.GetUserEmail(repo)
		err = git.Config("user.name", user, &git.Options{
			Path: repo.Path,
		})
		if err != nil {
			return err
		}

		err = git.Config("user.email", email, &git.Options{
			Path: repo.Path,
		})
		if err != nil {
			return err
		}

	case err == nil:
		if !stat.IsDir() {
			return fmt.Errorf("repo %s: %s is not a directory", repo.FullName(), repo.Path)
		}

	default:
		return errors.Trace(err, "check repo exists")
	}

	if ctx.store.GetByName(remote.Name, repo.Name) == nil {
		return ctx.store.Add(repo)
	}
	return nil
}

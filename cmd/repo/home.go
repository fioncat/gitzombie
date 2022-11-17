package repo

import (
	"fmt"
	"os"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type HomeFlags struct {
	Search bool
}

var Home = app.Register(&app.Command[HomeFlags, Data]{
	Use:    "repo {remote} {repo}",
	Desc:   "Print the home path of a repo",
	Action: "Home",

	Init: initData[HomeFlags],

	Prepare: func(cmd *cobra.Command, flags *HomeFlags) {
		cmd.Args = cobra.RangeArgs(1, 2)
		cmd.ValidArgsFunction = app.Comp(compRemote, compRepo)
		cmd.Flags().BoolVarP(&flags.Search, "search", "s", false, "search from remote")
	},

	Run: func(ctx *app.Context[HomeFlags, Data]) error {
		var repo *core.Repository
		var err error
		if ctx.Flags.Search {
			repo, err = homeSearchRepo(ctx)
		} else {
			repo, err = getLocal(ctx, ctx.Arg(1))
		}
		if err != nil {
			return err
		}

		err = homeEnsureRepo(ctx, repo)
		if err != nil {
			return err
		}
		repo.View++
		fmt.Println(repo.Path)
		return nil

	},
})

func homeSearchRepo(ctx *app.Context[HomeFlags, Data]) (*core.Repository, error) {
	apiRepo, err := apiSearch(ctx, ctx.Arg(1))
	if err != nil {
		return nil, err
	}

	repo := ctx.Data.Store.GetByName(ctx.Data.Remote.Name, apiRepo.Name)
	if repo == nil {
		repo, err = core.CreateRepository(ctx.Data.Remote, apiRepo.Name)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}

func homeEnsureRepo(ctx *app.Context[HomeFlags, Data], repo *core.Repository) error {
	url, err := ctx.Data.Remote.GetCloneURL(repo)
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
		user, email := ctx.Data.Remote.GetUserEmail(repo)
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

	if ctx.Data.Store.GetByName(ctx.Data.Remote.Name, repo.Name) == nil {
		return ctx.Data.Store.Add(repo)
	}
	return nil
}

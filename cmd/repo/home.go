package repo

import (
	"fmt"
	"os"

	"github.com/fioncat/gitzombie/api"
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

var Home = app.Register(&app.Command[HomeFlags, core.RepositoryStorage]{
	Use:  "home {remote} {repo}",
	Desc: "Enter or clone a repo",

	Init: initData[HomeFlags],

	Prepare: func(cmd *cobra.Command, flags *HomeFlags) {
		cmd.Args = cobra.RangeArgs(1, 2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompRepo)
		cmd.Flags().BoolVarP(&flags.Search, "search", "s", false, "search from remote")
	},

	Run: func(ctx *app.Context[HomeFlags, core.RepositoryStorage]) error {
		var repo *core.Repository
		var err error
		remote, err := core.GetRemote(ctx.Arg(0))
		if err != nil {
			return err
		}

		if ctx.Flags.Search {
			repo, err = homeSearchRepo(ctx, remote)
		} else {
			repo, err = ctx.Data.GetLocal(remote, ctx.Arg(1))
		}
		if err != nil {
			return err
		}

		err = homeEnsureRepo(ctx, remote, repo)
		if err != nil {
			return err
		}
		repo.MarkAccess()
		fmt.Println(repo.Path)
		return nil

	},
})

func homeSearchRepo(ctx *app.Context[HomeFlags, core.RepositoryStorage], remote *core.Remote) (*core.Repository, error) {
	apiRepo, err := api.SearchRepo(remote, ctx.Arg(1))
	if err != nil {
		return nil, err
	}

	repo := ctx.Data.GetByName(remote.Name, apiRepo.Name)
	if repo == nil {
		repo, err = core.WorkspaceRepository(remote, apiRepo.Name)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}

func homeEnsureRepo(ctx *app.Context[HomeFlags, core.RepositoryStorage], remote *core.Remote, repo *core.Repository) error {
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

	if ctx.Data.GetByName(remote.Name, repo.Name) == nil {
		return ctx.Data.Add(repo)
	}
	return nil
}

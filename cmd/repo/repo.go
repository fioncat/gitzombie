package repo

import (
	"fmt"
	"os"
	"strings"

	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Context struct {
	store *core.RepositoryStorage
}

type App struct{}

func (app *App) Name() string { return "repo" }

func (app *App) BuildContext(args []string) (*Context, error) {
	ctx := new(Context)
	var err error
	ctx.store, err = core.NewRepositoryStorage()
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (app *App) Close(ctx *Context) error {
	err := ctx.store.Close()
	if err != nil {
		return errors.Trace(err, "save data")
	}
	return nil
}

func getRemote(name string) (*core.Remote, error) {
	if name == "" {
		return nil, errors.New("you must specifiy a remote")
	}
	return core.GetRemote(name)
}

func getLocalRepo(ctx *Context, remote *core.Remote, name string) (*core.Repository, error) {
	if strings.HasSuffix(name, "/") || name == "" {
		group := strings.Trim(name, "/")
		allRepos := ctx.store.List(remote.Name)
		var repos []*core.Repository
		var items []string
		for _, repo := range allRepos {
			if group != "" && repo.Group() != group {
				continue
			}
			repos = append(repos, repo)
			var item string
			if group != "" {
				item = repo.Base()
			} else {
				item = repo.Name
			}
			items = append(items, item)
		}
		idx, err := term.FuzzySearch("repo", items)
		if err != nil {
			return nil, errors.Trace(err, "fzf search")
		}
		return repos[idx], nil
	}

	var err error
	repo := ctx.store.GetByName(remote.Name, name)
	if repo == nil {
		repo, err = core.CreateRepository(remote, name)
		if err != nil {
			return nil, errors.Trace(err, "create repository")
		}
	}
	return repo, nil
}

func ensureRepo(ctx *Context, remote *core.Remote, repo *core.Repository) error {
	url, err := remote.GetCloneURL(repo)
	if err != nil {
		return errors.Trace(err, "get clone url")
	}
	stat, err := os.Stat(repo.Path)
	switch {
	case os.IsNotExist(err):
		term.ConfirmExit("repo %s does not exists, do you want to clone it to %s", repo.FullName(), repo.Path)
		err = term.Exec("git", "clone", url, repo.Path)
		if err != nil {
			return err
		}
		user, email := remote.GetUserEmail(repo)
		err = term.Exec("git", "-C", repo.Path, "config", "user.name", user)
		if err != nil {
			return err
		}

		err = term.Exec("git", "-C", repo.Path, "config", "user.email", email)
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

type Home struct{}

func (home *Home) Name() []string { return []string{"home", "repo"} }
func (home *Home) Desc() string   { return "print the home path of a repo" }

func (home *Home) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.RangeArgs(1, 2)
}

func (home *Home) Handle(ctx *Context, args common.Args) error {
	remote, err := getRemote(args.Get(0))
	if err != nil {
		return err
	}

	repo, err := getLocalRepo(ctx, remote, args.Get(1))
	if err != nil {
		return err
	}

	err = ensureRepo(ctx, remote, repo)
	if err != nil {
		return err
	}

	repo.View++
	fmt.Println(repo.Path)
	return nil
}

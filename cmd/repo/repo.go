package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/api/github"
	"github.com/fioncat/gitzombie/api/gitlab"
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var (
	providers    map[string]api.Provider
	providerLock sync.Mutex
)

func getProvider(remote *core.Remote) (api.Provider, error) {
	providerLock.Lock()
	defer providerLock.Unlock()
	p := providers[remote.Name]
	if p != nil {
		return p, nil
	}
	var err error
	switch remote.Provider {
	case "github":
		p, err = github.New(remote)

	case "gitlab":
		p, err = gitlab.New(remote)

	default:
		return nil, fmt.Errorf("unknown provider %s", remote.Provider)
	}
	return p, err
}

func execProvider(op string, remote *core.Remote, h func(p api.Provider) error) error {
	p, err := getProvider(remote)
	if err != nil {
		return err
	}

	term.PrintSearch("%s %s", op, p.Name())
	err = h(p)
	return errors.Trace(err, "request %s api", p.Name())
}

type Context struct {
	store *core.RepositoryStorage
}

type App struct{}

func (app *App) Name() string { return "repo" }

func (app *App) BuildContext(args common.Args) (*Context, error) {
	ctx := new(Context)
	var err error
	ctx.store, err = core.NewRepositoryStorage()
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (app *App) Ops() []common.Operation[Context] {
	return []common.Operation[Context]{
		&Home{}, &RemoteHome{},
		&List{}, &Attach{},
		&Delete{},
	}
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

func getRemoteRepo(remote *core.Remote, query string) (*api.Repository, error) {
	var group string
	if strings.HasSuffix(query, "/") {
		group = strings.Trim(query, "/")
		query = ""
	} else {
		group, query = core.SplitGroup(query)
	}

	var repos []*api.Repository
	var err error
	err = execProvider("Searching", remote, func(p api.Provider) error {
		repos, err = p.SearchRepositories(group, query)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, api.ErrNoResult
	}
	if len(repos) == 1 {
		return repos[0], nil
	}
	items := make([]string, len(repos))
	for i, repo := range repos {
		var item string
		if group != "" {
			_, item = core.SplitGroup(repo.Name)
		} else {
			item = repo.Name
		}
		items[i] = item
	}
	idx, err := term.FuzzySearch("repo", items)
	if err != nil {
		return nil, err
	}
	return repos[idx], nil
}

func ensureRepo(ctx *Context, remote *core.Remote, repo *core.Repository) error {
	url, err := remote.GetCloneURL(repo)
	if err != nil {
		return errors.Trace(err, "get clone url")
	}
	stat, err := os.Stat(repo.Path)
	switch {
	case os.IsNotExist(err):
		term.ConfirmExit("repo %s does not exists, do you want to clone it", repo.FullName())
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

func (home *Home) Use() string    { return "repo remote [repo]" }
func (home *Home) Desc() string   { return "print the home path of a repo" }
func (home *Home) Action() string { return "home" }

func (home *Home) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.RangeArgs(1, 2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compRepo)
}

func (home *Home) Run(ctx *Context, args common.Args) error {
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

type RemoteHome struct{}

func (home *RemoteHome) Use() string    { return "remote remote query" }
func (home *RemoteHome) Desc() string   { return "search remote and print the home path" }
func (home *RemoteHome) Action() string { return "home" }

func (home *RemoteHome) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compGroup)
}

func (home *RemoteHome) Run(ctx *Context, args common.Args) error {
	remote, err := getRemote(args.Get(0))
	if err != nil {
		return err
	}
	remoteRepo, err := getRemoteRepo(remote, args.Get(1))
	if err != nil {
		return err
	}

	repo := ctx.store.GetByName(remote.Name, remoteRepo.Name)
	if repo == nil {
		repo, err = core.CreateRepository(remote, remoteRepo.Name)
		if err != nil {
			return errors.Trace(err, "create repository")
		}
	}
	err = ensureRepo(ctx, remote, repo)
	if err != nil {
		return err
	}

	repo.View++
	fmt.Println(repo.Path)
	return nil
}

type Attach struct{}

func (attach *Attach) Use() string    { return "attach remote repo" }
func (attach *Attach) Desc() string   { return "attach current path to repo" }
func (attach *Attach) Action() string { return "" }

func (attach *Attach) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compGroup)
}

func (attach *Attach) Run(ctx *Context, args common.Args) error {
	dir, err := os.Getwd()
	if err != nil {
		return errors.Trace(err, "get current dir")
	}
	gitDir := filepath.Join(dir, ".git")
	stat, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return common.ErrNotGit
		}
		return errors.Trace(err, "check git exists")
	}
	if !stat.IsDir() {
		return common.ErrNotGit
	}

	remote, err := getRemote(args.Get(0))
	if err != nil {
		return err
	}
	remoteRepo, err := getRemoteRepo(remote, args.Get(1))
	if err != nil {
		return err
	}

	repo, err := core.AttachRepository(remote, remoteRepo.Name, dir)
	if err != nil {
		return err
	}

	err = ctx.store.Add(repo)
	if err != nil {
		return err
	}

	if term.Confirm("overwrite git url") {
		url, err := remote.GetCloneURL(repo)
		if err != nil {
			return err
		}
		err = term.Exec("git", "remote", "set-url", "origin", url)
		if err != nil {
			return err
		}
	}

	if term.Confirm("overwrite user and email") {
		user, email := remote.GetUserEmail(repo)
		err = term.Exec("git", "config", "user.name", user)
		if err != nil {
			return err
		}

		err = term.Exec("git", "config", "user.email", email)
		if err != nil {
			return err
		}
	}

	return nil
}

type List struct{}

func (list *List) Use() string    { return "repo [remote] [repo]" }
func (list *List) Desc() string   { return "list remotes or repos" }
func (list *List) Action() string { return "list" }

func (list *List) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.MaximumNArgs(2)
}

func (list *List) Run(ctx *Context, args common.Args) error {
	ctx.store.ReadOnly()
	remoteName := args.Get(0)
	if remoteName == "" {
		remoteNames, err := core.ListRemoteNames()
		if err != nil {
			return err
		}
		for _, name := range remoteNames {
			fmt.Println(name)
		}
		return nil
	}

	repos := ctx.store.List(remoteName)
	for _, repo := range repos {
		fmt.Println(repo.Name)
	}
	return nil
}

type Delete struct{}

func (d *Delete) Use() string    { return "repo remote repo" }
func (d *Delete) Desc() string   { return "delete a repo" }
func (d *Delete) Action() string { return "delete" }

func (d *Delete) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compRepo)
}

func (d *Delete) Run(ctx *Context, args common.Args) error {
	remote, err := getRemote(args.Get(0))
	if err != nil {
		return err
	}

	repo, err := getLocalRepo(ctx, remote, args.Get(1))
	if err != nil {
		return err
	}
	if !term.Confirm("delete %s", repo.Path) {
		return nil
	}
	_, err = os.Stat(repo.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		err = os.RemoveAll(repo.Path)
		if err != nil {
			return errors.Trace(err, "remove repo")
		}
	}

	ctx.store.Delete(repo)
	return nil
}

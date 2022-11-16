package repo

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/api/github"
	"github.com/fioncat/gitzombie/api/gitlab"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
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

	term.PrintOperation("calling %s API to %s", p.Name(), op)
	err = h(p)
	return errors.Trace(err, "request %s api", p.Name())
}

func getRemote(name string) (*core.Remote, error) {
	if name == "" {
		return nil, errors.New("you must specifiy a remote")
	}
	return core.GetRemote(name)
}

func getLocal(ctx *Context, name string) (*core.Repository, error) {
	if strings.HasSuffix(name, "/") || name == "" {
		group := strings.Trim(name, "/")
		allRepos := ctx.store.List(ctx.remote.Name)
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
	repo := ctx.store.GetByName(ctx.remote.Name, name)
	if repo == nil {
		repo, err = core.CreateRepository(ctx.remote, name)
		if err != nil {
			return nil, errors.Trace(err, "create repository")
		}
	}
	return repo, nil
}

func getCurrent(ctx *Context) (*core.Repository, error) {
	path, err := git.EnsureCurrent()
	if err != nil {
		return nil, err
	}

	repo, err := ctx.store.GetByPath(path)
	if err != nil {
		return nil, err
	}

	remote, err := getRemote(repo.Remote)
	if err != nil {
		return nil, err
	}

	ctx.remote = remote
	return repo, nil
}

func apiSearch(ctx *Context, query string) (*api.Repository, error) {
	if query == "" {
		return nil, errors.New("please provide query statement")
	}
	var group string
	var op string
	if strings.HasSuffix(query, "/") {
		group = strings.Trim(query, "/")
		query = ""
		op = fmt.Sprintf("search group %q", group)
	} else {
		op = fmt.Sprintf("search %q", query)
		group, query = core.SplitGroup(query)
	}

	var repos []*api.Repository
	var err error
	err = execProvider(op, ctx.remote, func(p api.Provider) error {
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

func apiGet(ctx *Context, repo *core.Repository) (*api.Repository, error) {
	var remoteRepo *api.Repository
	var err error
	err = execProvider("get repository info", ctx.remote, func(p api.Provider) error {
		remoteRepo, err = p.GetRepository(repo.Name)
		return err
	})
	return remoteRepo, err
}

func convertToGroups(repos []*core.Repository) []string {
	groups := make([]string, 0)
	groupMap := make(map[string]struct{})
	for _, repo := range repos {
		group, _ := core.SplitGroup(repo.Name)
		if _, ok := groupMap[group]; ok {
			continue
		}
		groupMap[group] = struct{}{}
		groups = append(groups, group+"/")
	}
	return groups
}

package api

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
)

type ProviderCreator func(*core.Remote) (Provider, error)

var creators = map[string]ProviderCreator{}

func Register(name string, creator ProviderCreator) {
	creators[name] = creator
}

var (
	providers    map[string]Provider = map[string]Provider{}
	providerLock sync.Mutex
)

func GetProvider(remote *core.Remote) (Provider, error) {
	providerLock.Lock()
	defer providerLock.Unlock()
	p := providers[remote.Name]
	if p != nil {
		return p, nil
	}
	creator := creators[remote.Provider]
	if creator == nil {
		return nil, fmt.Errorf("unknown provider %s", remote.Provider)
	}

	if remote.TokenSecret {
		key := remote.Token
		if key == "" {
			key = fmt.Sprintf("%s_token", remote.Name)
		}

		token, err := core.GetSecret(key, true)
		if err != nil {
			return nil, errors.Trace(err, "get secret token")
		}
		remote.Token = token
	} else {
		remote.Token = os.ExpandEnv(remote.Token)
	}

	var err error
	p, err = creator(remote)
	if err != nil {
		return nil, err
	}
	providers[remote.Name] = p
	return p, nil
}

func Exec(op string, remote *core.Remote, h func(p Provider) error) error {
	p, err := GetProvider(remote)
	if err != nil {
		return err
	}

	term.PrintOperation("calling %s API to %s", p.Name(), op)
	err = h(p)
	return errors.Trace(err, "request %s api", p.Name())
}

func SearchRepo(remote *core.Remote, query string) (*Repository, error) {
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

	var repos []*Repository
	var err error
	err = Exec(op, remote, func(p Provider) error {
		repos, err = p.SearchRepositories(group, query)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, ErrNoResult
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

func GetRepo(remote *core.Remote, repo *core.Repository) (*Repository, error) {
	var remoteRepo *Repository
	var err error
	err = Exec("get repository info", remote, func(p Provider) error {
		remoteRepo, err = p.GetRepository(repo.Name)
		return err
	})
	return remoteRepo, err
}

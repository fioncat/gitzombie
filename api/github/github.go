package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type Provider struct {
	ctx context.Context
	cli *github.Client

	remote *core.Remote
}

func New(remote *core.Remote) (api.Provider, error) {
	var httpCli *http.Client
	ctx := context.Background()
	if remote.Token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: remote.Token,
		})
		httpCli = oauth2.NewClient(ctx, ts)
	}

	cli := github.NewClient(httpCli)
	return &Provider{
		cli:    cli,
		ctx:    ctx,
		remote: remote,
	}, nil
}

func (p *Provider) Name() string { return "Github" }

func (p *Provider) SearchRepositories(group, query string) ([]*api.Repository, error) {
	if group != "" {
		return p.searchInGroup(group, query)
	}
	result, resp, err := p.cli.Search.Repositories(p.ctx,
		query, &github.SearchOptions{
			ListOptions: github.ListOptions{
				PerPage: config.Get().SearchLimit,
			},
		})
	if err = p.wrapResp(query, resp, err); err != nil {
		return nil, err
	}
	if len(result.Repositories) == 0 {
		return nil, api.ErrNoResult
	}

	repos := make([]*api.Repository, len(result.Repositories))
	for i, githubRepo := range result.Repositories {
		repo := p.convertRepo(githubRepo)
		repos[i] = repo
	}
	return repos, nil
}

func (p *Provider) GetRepository(name string) (*api.Repository, error) {
	owner, repoName, err := parseOwner(name)
	if err != nil {
		return nil, err
	}

	githubRepo, resp, err := p.cli.Repositories.Get(p.ctx, owner, repoName)
	if err = p.wrapResp(name, resp, err); err != nil {
		return nil, err
	}
	if githubRepo == nil {
		return nil, p.notFound(name)
	}
	return p.convertRepo(githubRepo), nil
}

func (p *Provider) searchInGroup(group, query string) ([]*api.Repository, error) {
	githubRepos, resp, err := p.cli.Repositories.List(p.ctx, group,
		&github.RepositoryListOptions{
			ListOptions: github.ListOptions{
				PerPage: config.Get().SearchLimit,
			},
		})
	if err = p.wrapResp(query, resp, err); err != nil {
		return nil, err
	}
	if len(githubRepos) == 0 {
		return nil, api.ErrNoResult
	}

	repos := make([]*api.Repository, 0, len(githubRepos))
	for _, githubRepo := range githubRepos {
		if query != "" {
			if !strings.Contains(githubRepo.GetFullName(), query) {
				continue
			}
		}
		repo := p.convertRepo(githubRepo)
		repos = append(repos, repo)
	}

	return repos, nil
}

func (p *Provider) wrapResp(name string, resp *github.Response, err error) error {
	if resp == nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return p.notFound(name)
	}
	return nil
}

func (p *Provider) notFound(name string) error {
	return fmt.Errorf("cannot find %q in Github", name)
}

func (p *Provider) convertRepo(githubRepo *github.Repository) *api.Repository {
	repo := &api.Repository{
		Name:   githubRepo.GetFullName(),
		Remote: p.remote,

		WebURL: githubRepo.GetHTMLURL(),

		DefaultBranch: githubRepo.GetDefaultBranch(),
	}
	if githubRepo.GetFork() && githubRepo.GetSource() != nil {
		forked := githubRepo.GetSource()
		repo.Upstream = p.convertRepo(forked)
	}
	return repo
}

func parseOwner(name string) (string, string, error) {
	tmp := strings.Split(name, "/")
	if len(tmp) != 2 {
		return "", "", fmt.Errorf("invalid Github repo name %q", name)
	}
	return tmp[0], tmp[1], nil
}

package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

func init() {
	api.Register("github", New)
}

type pullOptions struct {
	Owner string
	Name  string

	Base string
	Head string

	HeadOwner string
}

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

func (p *Provider) ListRepositories(group string) ([]*api.Repository, error) {
	var page int = 1
	var repos []*api.Repository
	for {
		githubRepos, _, err := p.cli.Repositories.List(p.ctx, group,
			&github.RepositoryListOptions{
				ListOptions: github.ListOptions{
					PerPage: config.Get().SearchLimit,
					Page:    page,
				},
			})
		if err != nil {
			return nil, err
		}
		page++

		if len(githubRepos) == 0 {
			return repos, nil
		}

		for _, githubRepo := range githubRepos {
			repo := p.convertRepo(githubRepo)
			repos = append(repos, repo)
		}
		time.Sleep(time.Millisecond * 50)
	}
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

func (p *Provider) GetMerge(repo *core.Repository, opts api.MergeOption) (string, error) {
	pullOpts, err := p.createPullOptions(repo, opts)
	if err != nil {
		return "", err
	}
	if opts.Upstream != nil {
		query := fmt.Sprintf("is:open is:pr author:%s head:%s base:%s repo:%s",
			pullOpts.HeadOwner, opts.SourceBranch, opts.TargetBranch, opts.Upstream.Name)
		result, resp, err := p.cli.Search.Issues(p.ctx, query, &github.SearchOptions{})
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if result.GetTotal() == 0 {
			return "", nil
		}

		return result.Issues[0].GetHTMLURL(), nil
	}

	prs, resp, err := p.cli.PullRequests.List(p.ctx, pullOpts.Owner, pullOpts.Name,
		&github.PullRequestListOptions{
			State: "open",
			Head:  pullOpts.Head,
			Base:  pullOpts.Base,
		})
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if len(prs) == 0 {
		return "", nil
	}
	pr := prs[0]
	return pr.GetHTMLURL(), nil
}

func (p *Provider) CreateMerge(repo *core.Repository, opts api.MergeOption) (string, error) {
	pullOpts, err := p.createPullOptions(repo, opts)
	if err != nil {
		return "", err
	}
	pr, _, err := p.cli.PullRequests.Create(p.ctx, pullOpts.Owner, pullOpts.Name,
		&github.NewPullRequest{
			Title: github.String(opts.Title),
			Body:  github.String(opts.Body),
			Head:  github.String(pullOpts.Head),
			Base:  github.String(pullOpts.Base),
		})
	if err != nil || pr == nil {
		return "", err
	}
	return pr.GetHTMLURL(), nil
}

func (p *Provider) createPullOptions(repo *core.Repository, opts api.MergeOption) (*pullOptions, error) {
	var targetRepo string
	var head string
	var headOwner string
	base := opts.TargetBranch
	if opts.Upstream != nil {
		// Create PR to upstream, the operation object is upstream itself.
		// The base is upstream targetBranch, The head is "user:sourceBranch".
		// For example, merge "fioncat:kubernetes" to "kubernetes:kubernetes"
		// Branch is "master", the params are:
		//   repo: kubernetes/kubernetes
		//   base: master
		//   head: fioncat:master
		targetRepo = opts.Upstream.Name
		headOwner = repo.Group()
		head = fmt.Sprintf("%s:%s", headOwner, opts.SourceBranch)
	} else {
		targetRepo = repo.Name
		head = opts.SourceBranch
	}
	owner, name, err := parseOwner(targetRepo)
	if err != nil {
		return nil, err
	}

	return &pullOptions{
		Owner: owner,
		Name:  name,
		Head:  head,
		Base:  base,

		HeadOwner: headOwner,
	}, nil
}

func (p *Provider) GetRelease(repo *core.Repository, tag string) (*api.Release, error) {
	var release *github.RepositoryRelease
	var resp *github.Response
	var err error
	if tag == "" {
		release, resp, err = p.cli.Repositories.GetLatestRelease(p.ctx,
			repo.Group(), repo.Base())
		if err = p.wrapResp("latest release", resp, err); err != nil {
			return nil, err
		}
	} else {
		release, resp, err = p.cli.Repositories.GetReleaseByTag(p.ctx,
			repo.Group(), repo.Base(), tag)
		if err = p.wrapResp("release "+tag, resp, err); err != nil {
			return nil, err
		}
	}
	return p.convertRelease(release), nil
}

func (p *Provider) ListReleases(repo *core.Repository) ([]*api.Release, error) {
	githubReleases, resp, err := p.cli.Repositories.ListReleases(p.ctx,
		repo.Group(), repo.Base(), &github.ListOptions{
			PerPage: config.Get().SearchLimit,
		})
	if err = p.wrapResp("releases", resp, err); err != nil {
		return nil, err
	}
	releases := make([]*api.Release, len(githubReleases))
	for i, githubRelease := range githubReleases {
		releases[i] = p.convertRelease(githubRelease)
	}
	return releases, nil
}

func (p *Provider) DownloadReleaseFile(repo *core.Repository, file *api.ReleaseFile) (io.ReadCloser, error) {
	rc, _, err := p.cli.Repositories.DownloadReleaseAsset(p.ctx,
		repo.Group(), repo.Base(), file.ID.(int64), http.DefaultClient)
	if err != nil {
		return nil, err
	}
	if rc == nil {
		return nil, api.ErrNoResult
	}
	return rc, nil
}

func (p *Provider) convertRelease(githubRelease *github.RepositoryRelease) *api.Release {
	release := &api.Release{
		Name:   githubRelease.GetName(),
		Tag:    githubRelease.GetTagName(),
		WebURL: githubRelease.GetHTMLURL(),
	}
	release.Files = make([]*api.ReleaseFile, len(githubRelease.Assets))
	for i, asset := range githubRelease.Assets {
		release.Files[i] = &api.ReleaseFile{
			ID:   asset.GetID(),
			Name: asset.GetName(),
			Size: int64(asset.GetSize()),
		}
	}
	return release
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

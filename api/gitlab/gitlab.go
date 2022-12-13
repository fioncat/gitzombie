package gitlab

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

func init() {
	api.Register("gitlab", New)
}

type Provider struct {
	cli    *gitlab.Client
	remote *core.Remote
}

func New(remote *core.Remote) (api.Provider, error) {
	url := remote.API
	if url == "" {
		url = fmt.Sprintf("https://%s/api/v4", remote.Host)
	}
	cli, err := gitlab.NewClient(remote.Token, gitlab.WithBaseURL(url))
	if err != nil {
		return nil, err
	}

	return &Provider{cli: cli, remote: remote}, nil
}

func (p *Provider) Name() string { return "Gitlab" }

func (p *Provider) SearchRepositories(group, query string) ([]*api.Repository, error) {
	var prjs []*gitlab.Project
	var err error
	var resp *gitlab.Response
	if group != "" {
		prjs, resp, err = p.cli.Groups.ListGroupProjects(group,
			&gitlab.ListGroupProjectsOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: config.Get().SearchLimit,
				},
				Search: gitlab.String(query),
			})
	} else {
		prjs, resp, err = p.cli.Projects.ListProjects(
			&gitlab.ListProjectsOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: config.Get().SearchLimit,
				},
				Search: gitlab.String(query),
			})
	}
	if err = p.wrapResp(query, resp, err); err != nil {
		return nil, err
	}
	if len(prjs) == 0 {
		return nil, api.ErrNoResult
	}

	repos := make([]*api.Repository, len(prjs))
	for i, prj := range prjs {
		repo, err := p.convertRepo(prj)
		if err != nil {
			return nil, err
		}
		repos[i] = repo
	}

	return repos, nil
}

func (p *Provider) ListRepositories(group string) ([]*api.Repository, error) {
	_, resp, err := p.cli.Groups.GetGroup(group, &gitlab.GetGroupOptions{
		WithProjects: gitlab.Bool(true),
	})
	if err = p.wrapResp(group, resp, err); err != nil {
		return nil, err
	}

	var page int = 1
	var repos []*api.Repository
	for {
		prjs, _, err := p.cli.Groups.ListGroupProjects(group, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: config.Get().SearchLimit,
			},
		})
		if err != nil {
			return nil, err
		}
		page++

		if len(prjs) == 0 {
			return repos, nil
		}

		for _, prj := range prjs {
			repo, err := p.convertRepo(prj)
			if err != nil {
				return nil, errors.Trace(err, "parse repo %q", prj.NameWithNamespace)
			}
			repos = append(repos, repo)
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func (p *Provider) GetRepository(name string) (*api.Repository, error) {
	prj, resp, err := p.cli.Projects.GetProject(name,
		&gitlab.GetProjectOptions{})
	if err = p.wrapResp(name, resp, err); err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, p.notFound(name)
	}

	return p.convertRepo(prj)
}

func (p *Provider) GetMerge(repo *core.Repository, opts api.MergeOption) (string, error) {
	// TODO: Support forked merge
	if opts.Upstream != nil {
		return "", errors.New("now we donot support merge to upstream repo")
	}
	mr, err := p.getMergeRequest(repo, opts)
	if err != nil {
		return "", err
	}
	if mr == nil {
		return "", nil
	}
	return mr.WebURL, nil
}

func (p *Provider) CreateMerge(repo *core.Repository, opts api.MergeOption) (string, error) {
	// TODO: Support forked merge
	if opts.Upstream != nil {
		return "", errors.New("now we donot support merge to upstream repo")
	}
	src, tar := opts.SourceBranch, opts.TargetBranch
	mr, _, err := p.cli.MergeRequests.CreateMergeRequest(repo.Name,
		&gitlab.CreateMergeRequestOptions{
			Title:        gitlab.String(opts.Title),
			SourceBranch: gitlab.String(src),
			TargetBranch: gitlab.String(tar),
		})
	if err != nil {
		return "", err
	}

	return mr.WebURL, nil
}

func (p *Provider) getMergeRequest(repo *core.Repository, opts api.MergeOption) (*gitlab.MergeRequest, error) {
	mrs, resp, err := p.cli.MergeRequests.ListProjectMergeRequests(repo.Name,
		&gitlab.ListProjectMergeRequestsOptions{
			State:        gitlab.String("opened"),
			SourceBranch: gitlab.String(opts.SourceBranch),
			TargetBranch: gitlab.String(opts.TargetBranch),
		})
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(mrs) == 0 {
		return nil, nil
	}
	return mrs[0], nil
}

func (p *Provider) GetRelease(repo *core.Repository, tag string) (*api.Release, error) {
	// TODO: support release
	return nil, errors.New("sorry, gitlab now does not support release")
}

func (p *Provider) ListReleases(repo *core.Repository) ([]*api.Release, error) {
	// TODO: support release
	return nil, errors.New("sorry, gitlab now does not support release")
}

func (p *Provider) DownloadReleaseFile(repo *core.Repository, file *api.ReleaseFile) (io.ReadCloser, error) {
	// TODO: support release
	return nil, errors.New("sorry, gitlab now does not support release")
}

func (p *Provider) wrapResp(name string, resp *gitlab.Response, err error) error {
	if resp == nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return p.notFound(name)
	}
	return nil
}

func (p *Provider) notFound(name string) error {
	return fmt.Errorf("cannot find %q in Gitlab", name)
}

func (p *Provider) convertRepo(prj *gitlab.Project) (*api.Repository, error) {
	repo := &api.Repository{
		Name: prj.PathWithNamespace,

		Remote: p.remote,

		WebURL: prj.WebURL,

		DefaultBranch: prj.DefaultBranch,

		Archived: prj.Archived,
	}
	if prj.ForkedFromProject != nil {
		forked := prj.ForkedFromProject
		forkedPrj, resp, err := p.cli.Projects.GetProject(forked.ID,
			&gitlab.GetProjectOptions{})
		if err = p.wrapResp(forked.NameWithNamespace, resp, err); err != nil {
			return nil, err
		}
		repo.Upstream, err = p.convertRepo(forkedPrj)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}

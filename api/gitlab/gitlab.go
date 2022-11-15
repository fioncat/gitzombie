package gitlab

import (
	"fmt"
	"net/http"

	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/xanzy/go-gitlab"
)

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
	return fmt.Errorf("cannot find %q in Github", name)
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

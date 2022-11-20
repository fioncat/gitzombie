package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/validate"
)

type Workflow struct {
	Select *WorkflowSelect `yaml:"select"`

	Jobs []*Job `yaml:"jobs" validate:"required,dive"`
}

type WorkflowSelect struct {
	Repos []string `yaml:"repos"`
	Dirs  []string `yaml:"dirs"`
}

func ListWorkflowNames() ([]string, error) {
	return listConfigObjects("workflows", yamlExt)
}

func GetWorkflow(name string) (*Workflow, error) {
	return getConfigObject("workflows", yamlExt, "workflow", name, func(w *Workflow) error {
		return validate.Do(w)
	})
}

type WorkflowMatchItem struct {
	Path string
	Env  osutil.Env

	Repo *Repository
}

func (s *WorkflowSelect) Match(store *RepositoryStorage) ([]*WorkflowMatchItem, error) {
	var items []*WorkflowMatchItem
	if len(s.Repos) > 0 {
		repos, err := s.matchRepos(store)
		if err != nil {
			return nil, err
		}
		for _, matchRepo := range repos {
			env := make(osutil.Env)
			var remote *Remote
			if matchRepo.Remote != "" {
				remote, err = GetRemote(matchRepo.Remote)
				if err != nil {
					return nil, errors.Trace(err, "get remote %q", matchRepo.Remote)
				}
			}
			err = matchRepo.SetEnv(remote, env)
			if err != nil {
				return nil, errors.Trace(err, "set repo env")
			}
			items = append(items, &WorkflowMatchItem{
				Path: matchRepo.Path,
				Env:  env,
				Repo: matchRepo,
			})
		}
	}
	if len(s.Dirs) > 0 {
		dirs, err := s.matchDirs()
		if err != nil {
			return nil, err
		}
		for _, dir := range dirs {
			items = append(items, &WorkflowMatchItem{
				Path: dir,
				Env:  make(osutil.Env),
			})
		}
	}
	uniqueItems := make([]*WorkflowMatchItem, 0, len(items))
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, ok := set[item.Path]; ok {
			continue
		}
		set[item.Path] = struct{}{}
		uniqueItems = append(uniqueItems, item)
	}
	return uniqueItems, nil
}

type workflowRepoMatch struct {
	remote  string
	pattern string
}

func (m *workflowRepoMatch) match(store *RepositoryStorage) ([]*Repository, error) {
	repos := store.List(m.remote)
	var filters []*Repository
	for _, repo := range repos {
		ok, err := filepath.Match(m.pattern, repo.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %v", m.pattern, err)
		}
		if ok {
			filters = append(filters, repo)
		}
	}
	return repos, nil
}

func (s *WorkflowSelect) matchRepos(store *RepositoryStorage) ([]*Repository, error) {
	repoMaches := make([]*workflowRepoMatch, len(s.Repos))
	for i, repoMatchStr := range s.Repos {
		var remoteName string
		var repoPattern string
		tmp := strings.Split(repoMatchStr, ":")
		switch len(tmp) {
		case 1:
			remoteName = tmp[0]
			repoPattern = "*"

		default:
			remoteName = tmp[0]
			repoPattern = strings.Join(tmp[1:], ":")
			if repoPattern == "" {
				repoPattern = "*"
			}
		}
		repoMaches[i] = &workflowRepoMatch{
			remote:  remoteName,
			pattern: repoPattern,
		}
	}

	var matchRepos []*Repository
	for _, repoMatch := range repoMaches {
		repos, err := repoMatch.match(store)
		if err != nil {
			return nil, err
		}
		matchRepos = append(matchRepos, repos...)
	}
	return matchRepos, nil
}

func (s *WorkflowSelect) matchDirs() ([]string, error) {
	var dirs []string
	for _, dir := range s.Dirs {
		var scan bool
		if strings.HasSuffix(dir, "*") {
			scan = true
			dir = strings.TrimSuffix(dir, "*")
		}
		dir = os.ExpandEnv(dir)
		exists, err := osutil.DirExists(dir)
		if err != nil {
			return nil, errors.Trace(err, "check dir exists")
		}
		if !exists {
			return nil, fmt.Errorf("%s is not exists", dir)
		}
		if !scan {
			err = git.EnsurePath(dir)
			if err != nil {
				if err == git.ErrNotGit {
					return nil, fmt.Errorf("%s is not a git repository", dir)
				}
				return nil, errors.Trace(err, "check git dir")
			}
			dirs = append(dirs, dir)
			continue
		}
		dir = strings.TrimSuffix(dir, "/")
		scanDir := strings.TrimSuffix(dir, "*")
		items, err := git.Discover(scanDir)
		if err != nil {
			return nil, errors.Trace(err, "scan git dir")
		}
		for _, item := range items {
			dirs = append(dirs, item.Path)
		}
	}
	return dirs, nil
}

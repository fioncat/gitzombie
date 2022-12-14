package core

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

func SplitGroup(name string) (string, string) {
	tmp := strings.Split(name, "/")
	if len(tmp) <= 1 {
		return "", name
	}
	group := filepath.Join(tmp[:len(tmp)-1]...)
	base := tmp[len(tmp)-1]
	return group, base
}

type Repository struct {
	Path string

	Name   string
	Remote string

	LastAccess int64
	Access     uint64

	workspace bool

	group string
	base  string

	score uint64
}

func WorkspaceRepository(remote *Remote, name string) (*Repository, error) {
	dir := config.Get().Workspace
	path := filepath.Join(dir, remote.Name, name)
	repo, err := AttachRepository(remote, name, path)
	if err != nil {
		return nil, err
	}
	repo.workspace = true
	return repo, nil
}

func AttachRepository(remote *Remote, name, path string) (*Repository, error) {
	name = strings.Trim(name, "/")
	repo := &Repository{
		Path:   path,
		Name:   name,
		Remote: remote.Name,
	}
	err := repo.normalize()
	if err != nil {
		return nil, errors.Trace(err, "normalize repository")
	}
	return repo, nil
}

func (repo *Repository) Group() string {
	return repo.group
}

func (repo *Repository) Base() string {
	return repo.base
}

func (repo *Repository) normalize() error {
	if repo.Path == "" || repo.Name == "" || repo.Remote == "" {
		return errors.New("repository data is invalid")
	}
	tmp := strings.Split(repo.Name, "/")
	if len(tmp) <= 1 {
		return fmt.Errorf("invalid repository name %q, missing group", repo.Name)
	}
	repo.group, repo.base = SplitGroup(repo.Name)
	if repo.group == "" {
		return fmt.Errorf("invalid repository name %q, missing group", repo.Name)
	}
	return nil
}

func (repo *Repository) FullName() string {
	return fmt.Sprintf("%s:%s", repo.Remote, repo.Name)
}

func (repo *Repository) Dir() string {
	return filepath.Dir(repo.Path)
}

func (repo *Repository) EnsureDir() (string, error) {
	dir := repo.Dir()
	err := osutil.EnsureDir(dir)
	return dir, err
}

func (repo *Repository) SetEnv(remote *Remote, env osutil.Env) error {
	env["REPO_NAME"] = repo.Name
	env["REPO_GROUP"] = repo.group
	env["REPO_BASE"] = repo.base
	env["REPO_REMOTE"] = repo.Remote
	env["REPO_PATH"] = repo.Path
	env["REPO_DIR"] = repo.Dir()

	if remote != nil {
		email, user := remote.GetUserEmail(repo)
		url, err := remote.GetCloneURL(repo)
		if err != nil {
			return errors.Trace(err, "get clone url")
		}

		env["REMOTE_EMAIL"] = email
		env["REMOTE_USER"] = user
		env["REMOTE_URL"] = url
	}
	return nil
}

func (repo *Repository) MarkAccess() {
	repo.Access++
	repo.LastAccess = time.Now().Unix()
}

var (
	hourFactor  uint64 = 16
	dayFactor   uint64 = 8
	weekFactor  uint64 = 2
	otherFactor uint64 = 1
)

// calculate score for a repo. The algorithm comes from:
//
//	https://github.com/ajeetdsouza/zoxide/wiki/Algorithm
//
// Each repo is assigned a access count, starting with 1 the
// first time it is accessed.
// Every repo access increases the access count by 1. When
// a query is made, we calculate score based on the last time
// the repo was accessed:
//   - Last access within the last hour: score = access * 16
//   - Last access within the last day:  score = access * 8
//   - Last access within the last week: score = access * 2
//   - Last access more that one week:   score = access
func (repo *Repository) Score() uint64 {
	now := time.Now().Unix()
	delta := now - repo.LastAccess
	if delta <= 0 {
		return 0
	}
	var factor uint64
	switch {
	case delta <= config.HourSeconds:
		factor = hourFactor

	case delta <= config.DaySeconds:
		factor = dayFactor

	case delta <= config.WeekSeconds:
		factor = weekFactor

	default:
		factor = otherFactor
	}
	return repo.Access * factor
}

func SortRepositories(repos []*Repository) {
	for _, repo := range repos {
		repo.score = repo.Score()
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].score > repos[j].score
	})
}

func NewLocalRepository(rootDir, name string) (*Repository, error) {
	name = strings.Trim(name, "/")
	group, base := SplitGroup(name)
	if group == "" || name == "" {
		return nil, fmt.Errorf("invalid repo name %q, must have a group", name)
	}
	path := filepath.Join(rootDir, name)

	return &Repository{
		Path:  path,
		Name:  name,
		group: group,
		base:  base,
	}, nil
}

func DiscoverLocalRepositories(rootDir string) ([]*Repository, error) {
	dirs, err := git.Discover(rootDir)
	if err != nil {
		return nil, err
	}
	repos := make([]*Repository, len(dirs))
	for i, dir := range dirs {
		repo, err := NewLocalRepository(rootDir, dir.Name)
		if err != nil {
			return nil, err
		}
		repos[i] = repo
	}
	return repos, nil
}

const repoStorageName = "repo"

type RepositoryStorage struct {
	repos []*Repository

	nameIndex map[string]map[string]*Repository
	pathIndex map[string]*Repository

	lock sync.RWMutex

	readonly bool
}

func NewRepositoryStorage() (*RepositoryStorage, error) {
	s := &RepositoryStorage{
		nameIndex: make(map[string]map[string]*Repository),
		pathIndex: make(map[string]*Repository),
	}
	err := s.init()
	if err != nil {
		return nil, errors.Trace(err, "init repository storage")
	}
	return s, nil
}

func (s *RepositoryStorage) init() error {
	path := config.GetLocalDir(repoStorageName)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Trace(err, "open data file")
	}
	defer file.Close()

	repos, err := s.read(file)
	if err != nil {
		return &parseRepositoryStorageError{
			path: path,
			err:  err,
		}
	}

	s.repos = repos
	if len(s.repos) == 0 {
		return nil
	}

	SortRepositories(repos)
	for _, repo := range s.repos {
		if _, ok := s.pathIndex[repo.Path]; ok {
			return &parseRepositoryStorageError{
				path: path,
				err:  fmt.Errorf("path %q is duplicate", path),
			}
		}
		s.pathIndex[repo.Path] = repo

		repoMap := s.nameIndex[repo.Remote]
		if repoMap == nil {
			repoMap = make(map[string]*Repository, 1)
			s.nameIndex[repo.Remote] = repoMap
		}
		if _, ok := repoMap[repo.Name]; ok {
			return &parseRepositoryStorageError{
				path: path,
				err:  fmt.Errorf("repo %s is duplicate", repo.FullName()),
			}
		}
		repoMap[repo.Name] = repo
	}

	return nil
}

func (s *RepositoryStorage) read(file *os.File) ([]*Repository, error) {
	decoder := gob.NewDecoder(file)
	var repos []*Repository
	err := decoder.Decode(&repos)
	if err != nil {
		return nil, errors.Trace(err, "decode repo data")
	}
	for _, repo := range repos {
		if repo.Path == "" {
			dir := config.Get().Workspace
			repo.Path = filepath.Join(dir, repo.Remote, repo.Name)
			repo.workspace = true
		}
		err = repo.normalize()
		if err != nil {
			return nil, errors.Trace(err, "normalize repo %s", repo.Name)
		}
	}
	return repos, nil
}

func (s *RepositoryStorage) List(remote string) []*Repository {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var repos []*Repository
	for _, repo := range s.repos {
		if remote != "" && repo.Remote != remote {
			continue
		}
		repos = append(repos, repo)
	}
	return repos
}

func (s *RepositoryStorage) readError(err error, field string) error {
	return fmt.Errorf("failed to read field %s: %v", field, err)
}

func (s *RepositoryStorage) Close() error {
	if s.readonly {
		return nil
	}
	path := config.GetLocalDir(repoStorageName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Trace(err, "open data file")
	}
	defer file.Close()

	return s.write(file)
}

func (s *RepositoryStorage) write(file *os.File) error {
	for _, repo := range s.repos {
		if repo.workspace {
			repo.Path = ""
		}
	}
	encoder := gob.NewEncoder(file)
	return errors.Trace(encoder.Encode(&s.repos), "encode repo")
}

func (s *RepositoryStorage) Add(repo *Repository) error {
	err := repo.normalize()
	if err != nil {
		return errors.Trace(err, "normalize repo")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if v := s.get(repo.Remote, repo.Name); v != nil {
		return fmt.Errorf("repo %s is already exists", repo.FullName())
	}

	if v := s.pathIndex[repo.Path]; v != nil {
		return fmt.Errorf("path %s is already bound to %s", repo.Path, v.FullName())
	}

	s.repos = append(s.repos, repo)
	repoMap := s.nameIndex[repo.Remote]
	if repoMap == nil {
		repoMap = make(map[string]*Repository, 1)
		s.nameIndex[repo.Remote] = repoMap
	}
	repoMap[repo.Name] = repo
	s.pathIndex[repo.Path] = repo

	return nil
}

func (s *RepositoryStorage) Delete(repo *Repository) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.repos) == 0 {
		return
	}

	newRepos := make([]*Repository, 0, len(s.repos)-1)
	for _, item := range s.repos {
		if item.Remote == repo.Remote && item.Name == repo.Name {
			continue
		}
		newRepos = append(newRepos, item)
	}
	s.repos = newRepos

	repoMap := s.nameIndex[repo.Remote]
	if repoMap != nil {
		delete(repoMap, repo.Name)
	}
	delete(s.pathIndex, repo.Path)
}

func (s *RepositoryStorage) GetByName(remote, name string) *Repository {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.get(remote, name)
}

func (s *RepositoryStorage) GetByPath(path string) (*Repository, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	repo := s.pathIndex[path]
	if repo == nil {
		return nil, fmt.Errorf("path %s does not attach to any repository, please use attach command to attach first", path)
	}
	return repo, nil
}

func (s *RepositoryStorage) get(remote, name string) *Repository {
	repoMap := s.nameIndex[remote]
	if repoMap == nil {
		return nil
	}
	return repoMap[name]
}

func (s *RepositoryStorage) ReadOnly() {
	s.readonly = true
}

func (s *RepositoryStorage) DeleteAll(repo *Repository) error {
	_, err := os.Stat(repo.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		err = os.RemoveAll(repo.Path)
		if err != nil {
			return errors.Trace(err, "remove dir for repo %q", repo.FullName())
		}
	}
	s.Delete(repo)
	return nil
}

func (s *RepositoryStorage) GetCurrent() (*Repository, error) {
	path, err := git.EnsureCurrent()
	if err != nil {
		return nil, err
	}

	return s.GetByPath(path)
}

func (s *RepositoryStorage) GetLocal(remote *Remote, name string) (*Repository, error) {
	if strings.HasSuffix(name, "/") || name == "" {
		group := strings.Trim(name, "/")
		allRepos := s.List(remote.Name)
		var repos []*Repository
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
	repo := s.GetByName(remote.Name, name)
	if repo == nil {
		repo, err = WorkspaceRepository(remote, name)
		if err != nil {
			return nil, errors.Trace(err, "create repository")
		}
	}
	return repo, nil
}

type parseRepositoryStorageError struct {
	path string
	err  error
}

func (err *parseRepositoryStorageError) Error() string {
	return err.err.Error()
}

func (err *parseRepositoryStorageError) Extra() {
	term.Println()
	term.Printf("The repository data is broken, please fix or delete it: %s", err.path)
}

func ConvertToGroups(repos []*Repository) []string {
	groups := make([]string, 0)
	groupMap := make(map[string]struct{})
	for _, repo := range repos {
		group, _ := SplitGroup(repo.Name)
		if _, ok := groupMap[group]; ok {
			continue
		}
		groupMap[group] = struct{}{}
		groups = append(groups, group+"/")
	}
	return groups
}

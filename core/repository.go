package core

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

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

	View uint64

	workspace bool

	group string
	base  string
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
	path := config.GetLocalDir("meta")

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
		return &readRepositoryError{
			path: path,
			err:  err,
		}
	}

	s.repos = repos
	if len(s.repos) == 0 {
		return nil
	}

	s.sort()
	for _, repo := range s.repos {
		if _, ok := s.pathIndex[repo.Path]; ok {
			return &readRepositoryError{
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
			return &readRepositoryError{
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
	path := config.GetLocalDir("meta")
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

func (s *RepositoryStorage) sort() {
	sort.Slice(s.repos, func(i, j int) bool {
		return s.repos[i].View > s.repos[j].View
	})
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

type readRepositoryError struct {
	path string
	err  error
}

func (err *readRepositoryError) Error() string {
	return err.err.Error()
}

func (err *readRepositoryError) Extra() {
	term.Print("")
	term.Print("yellow|The repository data is broken, please fix or delete it: %s|", err.path)
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

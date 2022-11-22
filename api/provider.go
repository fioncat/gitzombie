package api

import (
	"io"

	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
)

var (
	ErrNoResult = errors.New("no result from remote server")
)

type Repository struct {
	Name string

	Remote *core.Remote

	WebURL string

	DefaultBranch string

	Upstream *Repository

	Archived bool
}

type MergeOption struct {
	Title string
	Body  string

	SourceBranch string
	TargetBranch string

	Upstream *Repository
}

type Release struct {
	Name string
	Tag  string

	WebURL string

	Files []*ReleaseFile
}

type ReleaseFile struct {
	ID any

	Name string
}

type Provider interface {
	Name() string
	SearchRepositories(group, query string) ([]*Repository, error)
	ListRepositories(group string) ([]*Repository, error)
	GetRepository(name string) (*Repository, error)

	GetMerge(repo *core.Repository, opts MergeOption) (string, error)
	CreateMerge(repo *core.Repository, opts MergeOption) (string, error)

	GetRelease(repo *core.Repository, tag string) (*Release, error)
	ListReleases(repo *core.Repository) ([]*Release, error)
	DownloadReleaseFile(repo *core.Repository, file *ReleaseFile) (io.ReadCloser, error)
}

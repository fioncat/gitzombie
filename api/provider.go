package api

import (
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

type Provider interface {
	Name() string
	SearchRepositories(group, query string) ([]*Repository, error)
	GetRepository(name string) (*Repository, error)

	GetMerge(repo *core.Repository, opts MergeOption) (string, error)
	CreateMerge(repo *core.Repository, opts MergeOption) (string, error)
}

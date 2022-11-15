package api

import (
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
)

var ErrNoResult = errors.New("no result from remote server")

type Repository struct {
	Name string

	Remote *core.Remote

	WebURL string

	DefaultBranch string

	Upstream *Repository

	Archived bool
}

type Provider interface {
	Name() string
	SearchRepositories(group, query string) ([]*Repository, error)
}

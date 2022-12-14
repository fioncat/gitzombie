package core

import (
	"fmt"

	"github.com/fioncat/gitzombie/pkg/validate"
)

type Remote struct {
	Name string `toml:"-"`

	Host string `toml:"host" validate:"required"`

	Protocol string `toml:"protocol" validate:"enum_protocol"`

	User  string `toml:"user" validate:"required"`
	Email string `toml:"email" validate:"email"`

	Provider    string `toml:"provider" validate:"enum_provider"`
	Token       string `toml:"token"`
	TokenSecret bool   `toml:"token_secret"`
	API         string `toml:"api" validate:"omitempty,uri"`

	Groups []*RemoteGroup `toml:"groups" validate:"unique=Name,dive"`
}

type RemoteGroup struct {
	Name string `toml:"name" validate:"required"`

	Protocol string `toml:"protocol" validate:"omitempty,enum_protocol"`

	User  string `toml:"user"`
	Email string `toml:"email" validate:"omitempty,email"`
}

func GetRemote(name string) (*Remote, error) {
	return getConfigObject("remotes", tomlExt, "remote", name, func(remote *Remote) error {
		remote.Name = name
		return validate.Do(remote)
	})
}

func ListRemoteNames() ([]string, error) {
	return listConfigObjects("remotes", tomlExt)
}

func (r *Remote) GetCloneURL(repo *Repository) (string, error) {
	protocol := r.Protocol
	group := r.matchGroup(repo)
	if group != nil && group.Protocol != "" {
		protocol = group.Protocol
	}
	switch protocol {
	case "https":
		return fmt.Sprintf("https://%s/%s.git", r.Host, repo.Name), nil
	case "ssh":
		return fmt.Sprintf("git@%s:%s.git", r.Host, repo.Name), nil
	}
	return "", fmt.Errorf("invalid protocol %s", protocol)
}

func (r *Remote) GetUserEmail(repo *Repository) (string, string) {
	user, email := r.User, r.Email
	group := r.matchGroup(repo)
	if group != nil {
		if group.User != "" {
			user = group.User
		}
		if group.Email != "" {
			email = group.Email
		}
	}
	return user, email
}

func (r *Remote) matchGroup(repo *Repository) *RemoteGroup {
	for _, group := range r.Groups {
		if repo.Group() == group.Name {
			return group
		}
	}
	return nil
}

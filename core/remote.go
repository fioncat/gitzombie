package core

import (
	"fmt"

	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/validate"
)

type Remote struct {
	Name string `toml:"-"`

	Host string `toml:"host" validate:"required"`

	Protocol string `toml:"protocol" validate:"enum_protocol"`

	User  string `toml:"user" validate:"required"`
	Email string `toml:"email" validate:"email"`

	Provider string `toml:"provider" validate:"enum_provider"`
	Token    string `toml:"token"`
	API      string `toml:"api" validate:"omitempty,uri"`

	Groups []*RemoteGroup `toml:"groups" validate:"unique=Name,dive"`
}

type RemoteGroup struct {
	Name string `toml:"name" validate:"required"`

	Protocol string `toml:"protocol" validate:"omitempty,enum_protocol"`

	User  string `toml:"user"`
	Email string `toml:"email" validate:"omitempty,email"`
}

func GetRemote(name string) (*Remote, error) {
	return getConfigObject(name, "remotes", "remote", ".toml", func(remote *Remote) error {
		err := validate.Do(remote)
		if err != nil {
			return err
		}
		remote.Name = name
		return nil
	})
}

func ListRemoteNames() ([]string, error) {
	return listConfigObjectNames("remotes", ".toml")
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

func (r *Remote) Setenv() error {
	return osutil.Setenv(map[string]string{
		"REMOTE_NAME":     r.Name,
		"REMOTE_HOST":     r.Host,
		"REMOTE_PROTOCOL": r.Protocol,
	})
}

func (r *Remote) matchGroup(repo *Repository) *RemoteGroup {
	for _, group := range r.Groups {
		if repo.Group() == group.Name {
			return group
		}
	}
	return nil
}

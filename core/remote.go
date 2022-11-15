package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/validate"
	"github.com/pelletier/go-toml/v2"
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
	filename := fmt.Sprintf("%s.toml", name)
	path := config.BaseDir("remotes", filename)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot find remote %s", name)
		}
		return nil, errors.Trace(err, "read remote file")
	}

	var remote Remote
	err = toml.Unmarshal(data, &remote)
	if err != nil {
		return nil, errors.Trace(err, "parse remote file")
	}
	remote.Name = name

	err = validate.Do(&remote)
	if err != nil {
		return nil, errors.Trace(err, "validate remote configuration for %s", name)
	}

	return &remote, nil
}

func ListRemoteNames() ([]string, error) {
	dir := config.BaseDir("remotes")
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Trace(err, "read remotes")
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		ext := filepath.Ext(name)
		if ext == ".toml" {
			name = strings.TrimSuffix(name, ext)
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names, nil
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
	if group.User != "" {
		user = group.User
	}
	if group.Email != "" {
		email = group.Email
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

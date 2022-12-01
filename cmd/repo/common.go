package repo

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/git"
)

func initData[Flags any](ctx *app.Context[Flags, core.RepositoryStorage]) error {
	store, err := core.NewRepositoryStorage()
	if err != nil {
		return err
	}
	ctx.OnClose(func() error { return store.Close() })
	ctx.Data = store
	return nil
}

type CloneTask struct {
	Path string
	URL  string

	User  string
	Email string
}

func (task *CloneTask) Execute() error {
	err := git.Clone(task.URL, task.Path, git.Mute)
	if err != nil {
		return err
	}

	err = git.Config("user.name", task.User, &git.Options{
		QuietCmd:    true,
		QuietStderr: true,

		Path: task.Path,
	})
	if err != nil {
		return err
	}
	return git.Config("user.email", task.Email, &git.Options{
		QuietCmd:    true,
		QuietStderr: true,

		Path: task.Path,
	})
}

package repo

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Attach struct{}

func (attach *Attach) Use() string    { return "attach remote repo" }
func (attach *Attach) Desc() string   { return "attach current path to repo" }
func (attach *Attach) Action() string { return "" }

func (attach *Attach) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compGroup)
}

func (attach *Attach) Run(ctx *Context, args common.Args) error {
	dir, err := git.EnsureCurrent()
	if err != nil {
		return err
	}

	apiRepo, err := apiSearch(ctx, args.Get(1))
	if err != nil {
		return err
	}

	repo, err := core.AttachRepository(ctx.remote, apiRepo.Name, dir)
	if err != nil {
		return err
	}

	err = ctx.store.Add(repo)
	if err != nil {
		return err
	}

	if term.Confirm("overwrite git url") {
		url, err := ctx.remote.GetCloneURL(repo)
		if err != nil {
			return err
		}
		err = git.SetRemoteURL("origin", url, git.Default)
		if err != nil {
			return err
		}
	}

	if term.Confirm("overwrite user and email") {
		user, email := ctx.remote.GetUserEmail(repo)
		err = git.Config("user.name", user, git.Default)
		if err != nil {
			return err
		}

		err = git.Config("user.email", email, git.Default)
		if err != nil {
			return err
		}
	}

	return nil
}

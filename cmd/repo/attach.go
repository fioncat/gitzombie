package repo

import (
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var Attach = app.Register(&app.Command[app.Empty, core.RepositoryStorage]{
	Use:  "attach {remote} {repo}",
	Desc: "Attach current path to a repo",

	Init: initData[app.Empty],

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompGroup)
	},

	Run: func(ctx *app.Context[app.Empty, core.RepositoryStorage]) error {
		dir, err := git.EnsureCurrent()
		if err != nil {
			return err
		}

		remote, err := core.GetRemote(ctx.Arg(0))
		if err != nil {
			return err
		}

		apiRepo, err := api.SearchRepo(remote, ctx.Arg(1))
		if err != nil {
			return err
		}

		repo, err := core.AttachRepository(remote, apiRepo.Name, dir)
		if err != nil {
			return err
		}

		err = ctx.Data.Add(repo)
		if err != nil {
			return err
		}

		if term.Confirm("overwrite git url") {
			url, err := remote.GetCloneURL(repo)
			if err != nil {
				return err
			}
			err = git.SetRemoteURL("origin", url, git.Default)
			if err != nil {
				return err
			}
		}

		if term.Confirm("overwrite user and email") {
			user, email := remote.GetUserEmail(repo)
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
	},
})

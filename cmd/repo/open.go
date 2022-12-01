package repo

import (
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var Open = app.Register(&app.Command[app.Empty, core.RepositoryStorage]{
	Use:    "repo {remote} {repo}",
	Desc:   "Open repo in default browser",
	Action: "Open",

	Init: initData[app.Empty],

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.MaximumNArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompRepo)
	},

	Run: func(ctx *app.Context[app.Empty, core.RepositoryStorage]) error {
		ctx.Data.ReadOnly()
		var apiRepo *api.Repository
		var err error
		switch ctx.ArgLen() {
		case 0:
			var repo *core.Repository
			repo, err = ctx.Data.GetCurrent()
			if err != nil {
				return err
			}
			remote, err := core.GetRemote(repo.Remote)
			if err != nil {
				return err
			}

			apiRepo, err = api.GetRepo(remote, repo)
			if err != nil {
				return err
			}

		default:
			remote, err := core.GetRemote(ctx.Arg(0))
			if err != nil {
				return err
			}

			apiRepo, err = api.SearchRepo(remote, ctx.Arg(1))
			if err != nil {
				return err
			}
		}

		return term.Open(apiRepo.WebURL)
	},
})

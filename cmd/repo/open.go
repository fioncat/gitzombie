package repo

import (
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var Open = app.Register(&app.Command[app.Empty, Data]{
	Use:    "repo {remote} {repo}",
	Desc:   "Open repo in default browser",
	Action: "Open",

	Init: initData[app.Empty],

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.MaximumNArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompRepo)
	},

	Run: func(ctx *app.Context[app.Empty, Data]) error {
		ctx.Data.Store.ReadOnly()
		var apiRepo *api.Repository
		var err error
		switch ctx.ArgLen() {
		case 0:
			repo, err := getCurrent(ctx)
			if err != nil {
				return err
			}
			apiRepo, err = apiGet(ctx, repo)
			if err != nil {
				return err
			}

		default:
			apiRepo, err = apiSearch(ctx, ctx.Arg(1))
			if err != nil {
				return err
			}
		}

		return term.Open(apiRepo.WebURL)
	},
})

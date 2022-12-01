package repo

import (
	"time"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var Describe = app.Register(&app.Command[app.Empty, core.RepositoryStorage]{
	Use:    "repo [remote] [repo]",
	Desc:   "Describe repo",
	Action: "Describe",

	Init: initData[app.Empty],

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.MaximumNArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompRepo)
	},

	Run: func(ctx *app.Context[app.Empty, core.RepositoryStorage]) error {
		ctx.Data.ReadOnly()
		var repo *core.Repository
		var err error
		var remote *core.Remote
		switch ctx.ArgLen() {
		case 0:
			repo, err = ctx.Data.GetCurrent()
			if err != nil {
				return err
			}
			remote, err = core.GetRemote(repo.Remote)
			if err != nil {
				return err
			}
		default:
			remote, err = core.GetRemote(ctx.Arg(0))
			if err != nil {
				return err
			}

			repo, err = ctx.Data.GetLocal(remote, ctx.Arg(1))
			if err != nil {
				return err
			}
		}

		timeStr := time.Unix(repo.LastAccess, 0).Format("2006-01-02 15:04:05")

		term.Print("Name:   green|%s|", repo.Name)
		term.Print("Group:  green|%s|", repo.Group())
		term.Print("Base:   green|%s|", repo.Base())
		term.Print("Remote: green|%s|", repo.Remote)

		term.Print("")
		term.Print("Access:")
		term.Print("* Count: green|%d|", repo.Access)
		term.Print("* Last:  green|%s|", timeStr)
		term.Print("* Score: green|%d|", repo.Score())
		return nil
	},
})

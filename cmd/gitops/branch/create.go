package branch

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/git"
)

var Create = app.Register(&app.Command[gitops.CreateFlags, app.Empty]{
	Use:    "branch {name}",
	Desc:   "Create a branch",
	Action: "Create",

	Prepare: gitops.PrepareCreate,

	Run: func(ctx *app.Context[gitops.CreateFlags, app.Empty]) error {
		name := ctx.Arg(0)
		err := git.Checkout(name, true, git.Default)
		if err != nil {
			return err
		}
		if !ctx.Flags.NoPush {
			err = git.Exec([]string{"push", "--set-upstream", ctx.Flags.Remote, name}, git.Default)
			if err != nil {
				return err
			}
		}
		return nil
	},
})

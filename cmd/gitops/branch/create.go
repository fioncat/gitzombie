package branch

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	NoPush bool
	Remote string
}

var Create = app.Register(&app.Command[CreateFlags, struct{}]{
	Use:    "branch {name}",
	Desc:   "Create a branch",
	Action: "Create",

	Prepare: func(cmd *cobra.Command, flags *CreateFlags) {
		cmd.Flags().BoolVarP(&flags.NoPush, "no-push", "", false, "donot push to remote")
		cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "origin", "remote name")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompGitRemote))

		cmd.Args = cobra.ExactArgs(1)
	},

	Run: func(ctx *app.Context[CreateFlags, struct{}]) error {
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

package tools

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/spf13/cobra"
)

type ResetFlags struct {
	Remote   string
	Upstream bool
}

var Reset = app.Register(&app.Command[ResetFlags, app.Empty]{
	Use:  "reset [-r remote] [-u] [target]",
	Desc: "Reset current branch to target",

	Prepare: func(cmd *cobra.Command, flags *ResetFlags) {
		cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "", "remote name")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompGitRemote))

		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompGitLocalBranch(false))

		cmd.Flags().BoolVarP(&flags.Upstream, "upstream", "u", false, "use upstream")
	},

	Init: func(_ *app.Context[ResetFlags, app.Empty]) error {
		err := git.EnsureNoUncommitted(git.Default)
		if err != nil {
			return err
		}
		return nil
	},

	Run: func(ctx *app.Context[ResetFlags, app.Empty]) error {
		branch := ctx.Arg(0)
		target, _, err := getTarget(branch, ctx.Flags.Remote, ctx.Flags.Upstream)
		if err != nil {
			return err
		}

		err = git.Pull(git.Default)
		if err != nil {
			return err
		}

		return git.Exec([]string{"reset", "--hard", target}, git.Default)
	},
})

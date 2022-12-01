package tools

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type RebaseFlags struct {
	Remote   string
	Push     bool
	Upstream bool
}

var Rebase = app.Register(&app.Command[RebaseFlags, app.Empty]{
	Use:  "rebase [-r remote] [target]",
	Desc: "Pull and rebase",

	Prepare: func(cmd *cobra.Command, flags *RebaseFlags) {
		cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "", "remote name")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompGitRemote))

		cmd.Flags().BoolVarP(&flags.Push, "push", "p", false, "push changes to remote")

		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompGitLocalBranch(false))

		cmd.Flags().BoolVarP(&flags.Upstream, "upstream", "u", false, "use upstream")
	},

	Init: func(_ *app.Context[RebaseFlags, app.Empty]) error {
		err := git.EnsureNoUncommitted(git.Default)
		if err != nil {
			return err
		}
		return nil
	},

	Run: func(ctx *app.Context[RebaseFlags, app.Empty]) error {
		branch := ctx.Arg(0)
		target, _, err := getTarget(branch, ctx.Flags.Remote, ctx.Flags.Upstream)
		if err != nil {
			return err
		}

		err = git.Pull(git.Default)
		if err != nil {
			return err
		}

		err = git.Exec([]string{"rebase", target}, git.Default)
		if err != nil {
			term.Warn("rebase: found conflict(s), please handle manually")
			return nil
		}

		if ctx.Flags.Push {
			return git.Exec([]string{"push", "-f"}, git.Default)
		}
		return nil
	},
})

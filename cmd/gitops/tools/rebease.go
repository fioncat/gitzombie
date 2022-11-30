package tools

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type RebaseFlags struct {
	Remote string
	Push   bool
}

var Rebase = app.Register(&app.Command[RebaseFlags, app.Empty]{
	Use:  "rebase [-r remote] [target]",
	Desc: "Pull and rebase",

	Prepare: func(cmd *cobra.Command, flags *RebaseFlags) {
		cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "origin", "remote name")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompGitRemote))

		cmd.Flags().BoolVarP(&flags.Push, "push", "p", false, "push changes to remote")

		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompGitLocalBranch(false))
	},

	Init: func(_ *app.Context[RebaseFlags, app.Empty]) error {
		err := git.EnsureNoUncommitted(git.Default)
		if err != nil {
			return err
		}
		return nil
	},

	Run: func(ctx *app.Context[RebaseFlags, app.Empty]) error {
		target := ctx.Arg(0)
		if target == "" {
			mainBranch, err := git.GetMainBranch(ctx.Flags.Remote, git.Default)
			if err != nil {
				return err
			}
			target = mainBranch
			term.Print("use target green|%s|", target)
		}

		err := git.Pull(git.Default)
		if err != nil {
			return err
		}
		err = git.Checkout(target, false, git.Default)
		if err != nil {
			return err
		}
		err = git.Pull(git.Default)
		if err != nil {
			return err
		}

		err = git.Checkout("-", false, git.Default)
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

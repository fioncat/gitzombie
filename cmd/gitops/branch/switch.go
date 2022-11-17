package branch

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var Switch = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "branch [branch]",
	Desc:   "Switch to a local branch",
	Action: "Switch",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompGitLocalBranch(false))
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		branch := ctx.Arg(0)
		if branch == "" {
			names, err := git.ListLocalBranchNames(false, git.Default)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return errors.New("no branch to search")
			}
			idx, err := term.FuzzySearch("branch", names)
			if err != nil {
				return err
			}
			branch = names[idx]
		}
		return git.Checkout(branch, false, git.Default)
	},
})

type SwitchRemoteFlags struct {
	gitops.RemoteFlags
	gitops.FetchFlags

	Local string
}

var SwitchRemote = app.Register(&app.Command[SwitchRemoteFlags, app.Empty]{
	Use:    "remote [-r remote] [--no-fetch] [--local {local}] {remote}",
	Desc:   "Switch to a remote branch",
	Action: "Switch",

	Prepare: func(cmd *cobra.Command, flags *SwitchRemoteFlags) {
		gitops.PreapreRemoteFlags(cmd, &flags.RemoteFlags)
		gitops.PreapreFetchFlags(cmd, &flags.FetchFlags)
		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompGitRemoteBranch)
		cmd.Flags().StringVarP(&flags.Local, "local", "", "", "local branch name")
	},

	Run: func(ctx *app.Context[SwitchRemoteFlags, app.Empty]) error {
		branch := ctx.Arg(0)
		if branch == "" {
			if !ctx.Flags.NoFetch {
				err := git.Fetch(ctx.Flags.Remote, true, false, git.Default)
				if err != nil {
					return err
				}
			}
			branches, err := git.ListLocalBranches(git.Default)
			if err != nil {
				return err
			}
			names, err := git.ListRemoteBranches(ctx.Flags.Remote, branches, git.Default)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return errors.New("no remote branch")
			}
			idx, err := term.FuzzySearch("remote branch", names)
			if err != nil {
				return err
			}
			branch = names[idx]
		}

		local := ctx.Flags.Local
		if local == "" {
			local = branch
		}

		target := fmt.Sprintf("%s/%s", ctx.Flags.Remote, branch)
		return git.Switch(local, target, git.Default)
	},
})

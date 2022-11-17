package branch

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type syncTask struct {
	branch string
	cmd    []string
	desc   string
}

type SyncFlags struct {
	NoDelete bool
	Remote   string
}

type SyncData struct {
	MainBranch    string
	BackupBranch  string
	CurrentBranch string

	Branches []*git.BranchDetail

	Tasks []*syncTask
}

var Sync = app.Register(&app.Command[SyncFlags, SyncData]{
	Use:    "branch",
	Desc:   "Sync branch with remote",
	Action: "Sync",

	Prepare: func(cmd *cobra.Command, flags *SyncFlags) {
		cmd.Flags().BoolVarP(&flags.NoDelete, "no-delete", "", false, "donot delete any branch")
		cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "origin", "remote name")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(gitops.CompRemote))
	},

	Init: func(ctx *app.Context[SyncFlags, SyncData]) error {
		ctx.Data = new(SyncData)
		err := git.EnsureNoUncommitted(git.Default)
		if err != nil {
			return err
		}
		err = git.Fetch(ctx.Flags.Remote, true, false, git.Default)
		if err != nil {
			return err
		}
		branches, err := git.ListLocalBranches(git.Default)
		if err != nil {
			return err
		}
		ctx.Data.Branches = branches

		branchWord := english.PluralWord(len(branches), "branch", "")
		term.Print("found magenta|%d| %s", len(branches), branchWord)

		mainBranch, err := git.GetMainBranch(ctx.Flags.Remote, git.Default)
		if err != nil {
			return err
		}
		ctx.Data.MainBranch = mainBranch
		term.Print("main branch is magenta|%s|", mainBranch)

		ctx.Data.BackupBranch = mainBranch
		err = syncCreateTasks(ctx)
		if err != nil {
			return err
		}
		if ctx.Data.CurrentBranch == "" {
			ctx.Data.CurrentBranch, err = git.GetCurrentBranch(git.Default)
			if err != nil {
				return err
			}
		}
		term.Print("backup branch is magenta|%s|", ctx.Data.BackupBranch)
		term.Print("")
		return nil
	},

	Run: func(ctx *app.Context[SyncFlags, SyncData]) error {
		if len(ctx.Data.Tasks) == 0 {
			term.Print("nothing to do")
			return nil
		}

		taskWord := english.Plural(len(ctx.Data.Tasks), "task", "")
		term.Print("we have %s to run:", taskWord)
		for _, task := range ctx.Data.Tasks {
			term.Print(task.desc)
		}
		term.ConfirmExit("continue")

		for _, task := range ctx.Data.Tasks {
			if ctx.Data.CurrentBranch != task.branch {
				git.Checkout(task.branch, false, git.QuietOutput)
				ctx.Data.CurrentBranch = task.branch
			}
			err := git.Exec(task.cmd, git.Default)
			if err != nil {
				return err
			}
		}

		if ctx.Data.CurrentBranch != ctx.Data.BackupBranch {
			return git.Checkout(ctx.Data.BackupBranch, false, git.QuietOutput)
		}
		return nil
	},
})

func syncCreateTasks(ctx *app.Context[SyncFlags, SyncData]) error {
	var tasks []*syncTask
	for _, branch := range ctx.Data.Branches {
		if branch.Current {
			if ctx.Flags.NoDelete || branch.RemoteStatus != git.RemoteStatusGone {
				ctx.Data.BackupBranch = branch.Name
			}
			ctx.Data.CurrentBranch = branch.Name
		}
		var desc string
		var cmd []string
		var tar string
		switch branch.RemoteStatus {
		case git.RemoteStatusAhead:
			tar = branch.Name
			desc = "green|push|  "
			cmd = []string{"push"}

		case git.RemoteStatusBehind:
			tar = branch.Name
			desc = "green|pull|  "
			cmd = []string{"pull"}

		case git.RemoteStatusGone:
			if ctx.Flags.NoDelete {
				continue
			}
			if branch.Name == ctx.Data.MainBranch {
				continue
			}
			tar = ctx.Data.MainBranch
			desc = "red|delete|"
			cmd = []string{"branch", "-D", branch.Name}

		default:
			continue
		}
		desc = fmt.Sprintf("  * %s %s", desc, branch.Name)
		tasks = append(tasks, &syncTask{
			branch: tar,
			cmd:    cmd,
			desc:   desc,
		})
	}
	return nil
}

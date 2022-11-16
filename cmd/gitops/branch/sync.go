package branch

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Sync struct {
	noDelete bool
	remote   string

	tasks []struct {
		branch string
		cmds   [][]string
		desc   string
	}
}

func (b *Sync) Use() string    { return "branch" }
func (b *Sync) Desc() string   { return "Sync branch with remote" }
func (b *Sync) Action() string { return "sync" }

func (b *Sync) Prepare(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&b.noDelete, "no-delete", "", false, "donot delete any branch")
	cmd.Flags().StringVarP(&b.remote, "remote", "r", "origin", "remote name")
	cmd.RegisterFlagCompletionFunc("remote", common.Comp(gitops.CompRemote))
}

func (b *Sync) Run(_ *struct{}, args common.Args) error {
	err := git.EnsureNoUncommitted(git.Default)
	if err != nil {
		return err
	}
	err = git.Fetch(b.remote, true, false, git.Default)
	if err != nil {
		return err
	}
	branches, err := git.ListLocalBranches(git.Default)
	if err != nil {
		return err
	}
	branchWord := english.PluralWord(len(branches), "branch", "")
	term.Print("found magenta|%d| %s", len(branches), branchWord)

	mainBranch, err := git.GetMainBranch(b.remote, git.Default)
	if err != nil {
		return err
	}
	term.Print("main branch is magenta|%s|", mainBranch)
	backupBranch := mainBranch

	for _, branch := range branches {
		if branch.Current {
			if b.noDelete || branch.RemoteStatus != git.RemoteStatusGone {
				backupBranch = branch.Name
			}
		}
		var desc string
		var ops [][]string
		var tar string
		switch branch.RemoteStatus {
		case git.RemoteStatusAhead:
			tar = branch.Name
			desc = "green|push|  "
			ops = [][]string{
				{"checkout", branch.Name},
				{"push"},
			}

		case git.RemoteStatusBehind:
			tar = branch.Name
			desc = "green|pull|  "
			ops = [][]string{
				{"checkout", branch.Name},
				{"pull"},
			}

		case git.RemoteStatusGone:
			if b.noDelete {
				continue
			}
			tar = mainBranch
			desc = "red|delete|"
			ops = [][]string{
				{"checkout", mainBranch},
				{"branch", "-D", branch.Name},
			}

		default:
			continue
		}
		desc = fmt.Sprintf("  * %s %s", desc, branch.Name)
		b.tasks = append(b.tasks, struct {
			branch string
			cmds   [][]string
			desc   string
		}{
			branch: tar,
			cmds:   ops,
			desc:   desc,
		})
	}
	term.Print("backup branch is magenta|%s|", backupBranch)
	term.Print("")
	if len(b.tasks) == 0 {
		term.Print("nothing to do")
		return nil
	}

	taskWord := english.Plural(len(b.tasks), "task", "")
	term.Print("we have %s to run:", taskWord)
	for _, task := range b.tasks {
		term.Print(task.desc)
	}
	term.ConfirmExit("continue")

	var cur string
	for _, task := range b.tasks {
		for _, cmds := range task.cmds {
			err = git.Exec(cmds, git.QuietOutput)
			if err != nil {
				return err
			}
		}
		cur = task.branch
	}

	if cur != backupBranch {
		return git.Checkout(backupBranch, false, git.QuietOutput)
	}
	return nil
}
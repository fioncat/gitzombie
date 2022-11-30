package tools

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type SquashFlags struct {
	Remote string

	Message string

	Num int
}

type SquashData struct {
	Word    string
	Commits []string
}

var Squash = app.Register(&app.Command[SquashFlags, SquashData]{
	Use:  "squash [-r remote] [-n num] [-m message] [target]",
	Desc: "Squash multiple commits into one",

	Prepare: func(cmd *cobra.Command, flags *SquashFlags) {
		cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "origin", "remote name")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompGitRemote))

		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompGitLocalBranch(false))

		cmd.Flags().IntVarP(&flags.Num, "num", "n", 0, "number of commits to squash")

		cmd.Flags().StringVarP(&flags.Message, "message", "m", "", "commit message")
	},

	Init: func(ctx *app.Context[SquashFlags, SquashData]) error {
		target := ctx.Arg(0)
		if target == "" {
			mainBranch, err := git.GetMainBranch(ctx.Flags.Remote, git.Default)
			if err != nil {
				return err
			}
			target = mainBranch
			term.Print("use target green|%s|", target)
		}

		commits, err := git.ListCommitsBetween(target, git.Default)
		if err != nil {
			return err
		}

		if len(commits) == 0 {
			return errors.New("no commit to squash")
		}
		if len(commits) == 1 {
			return errors.New("only found one commit, no need to squash")
		}

		commitWord := english.Plural(len(commits), "commit", "")
		if ctx.Flags.Num > 1 {
			if ctx.Flags.Num > len(commits) {
				return fmt.Errorf("%s between HEAD and %s, but number %d is too big", commitWord, target, ctx.Flags.Num)
			}
			commits = commits[:ctx.Flags.Num]
		}

		ctx.Data = &SquashData{
			Word:    commitWord,
			Commits: commits,
		}
		return nil
	},

	Run: func(ctx *app.Context[SquashFlags, SquashData]) error {
		commits := ctx.Data.Commits
		if len(commits) == 0 {
			term.Print("nothing to do")
			return nil
		}

		term.Print("")
		term.Print("found green|%s| to squash:", ctx.Data.Word)
		for _, commit := range commits {
			term.Print("* %s", commit)
		}
		term.ConfirmExit("continue")

		num := len(commits)
		err := git.Exec([]string{
			"reset", "--soft", fmt.Sprintf("HEAD~%d", num),
		}, git.Default)
		if err != nil {
			return err
		}

		commitArgs := []string{"commit"}
		if ctx.Flags.Message != "" {
			commitArgs = append(commitArgs, "-m", ctx.Flags.Message)
		}
		cmd := exec.Command("git", commitArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err = cmd.Run()
		return errors.Trace(err, "git commit")
	},
})

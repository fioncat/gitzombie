package branch

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/spf13/cobra"
)

type Create struct {
	noPush bool
	remote string
}

func (b *Create) Use() string    { return "branch" }
func (b *Create) Desc() string   { return "Create a branch" }
func (b *Create) Action() string { return "create" }

func (b *Create) Prepare(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&b.noPush, "no-push", "", false, "donot push to remote")
	cmd.Flags().StringVarP(&b.remote, "remote", "r", "origin", "remote name")
	cmd.RegisterFlagCompletionFunc("remote", common.Comp(gitops.CompRemote))

	cmd.Args = cobra.ExactArgs(1)
}

func (b *Create) Run(_ *struct{}, args common.Args) error {
	name := args.Get(0)
	err := git.Checkout(name, true, git.Default)
	if err != nil {
		return err
	}
	if !b.noPush {
		err = git.Exec([]string{"push", "--set-upstream", b.remote, name}, git.Default)
		if err != nil {
			return err
		}
	}
	return nil
}

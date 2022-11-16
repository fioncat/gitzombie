package repo

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

type List struct {
	group bool
}

func (list *List) Use() string    { return "repo [remote] [repo]" }
func (list *List) Desc() string   { return "list remotes or repos" }
func (list *List) Action() string { return "list" }

func (list *List) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.MaximumNArgs(2)
	cmd.Flags().BoolVarP(&list.group, "group", "", false, "list group")
}

func (list *List) Run(ctx *Context, args common.Args) error {
	ctx.store.ReadOnly()
	remoteName := args.Get(0)
	if remoteName == "" {
		remoteNames, err := core.ListRemoteNames()
		if err != nil {
			return err
		}
		for _, name := range remoteNames {
			fmt.Println(name)
		}
		return nil
	}

	repos := ctx.store.List(remoteName)
	if list.group {
		groups := convertToGroups(repos)
		for _, group := range groups {
			fmt.Println(group)
		}
		return nil
	}
	for _, repo := range repos {
		fmt.Println(repo.Name)
	}
	return nil
}

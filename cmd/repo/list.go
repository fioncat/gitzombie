package repo

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

type ListFlags struct {
	Group bool
}

var List = app.Register(&app.Command[ListFlags, Data]{
	Use:    "repo {remote} {repo}",
	Desc:   "list remotes or repos",
	Action: "List",

	Init: initData[ListFlags],

	Prepare: func(cmd *cobra.Command, flags *ListFlags) {
		cmd.Args = cobra.MaximumNArgs(2)
		cmd.Flags().BoolVarP(&flags.Group, "group", "", false, "list group")
	},

	Run: func(ctx *app.Context[ListFlags, Data]) error {
		ctx.Data.Store.ReadOnly()
		remoteName := ctx.Arg(0)
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

		repos := ctx.Data.Store.List(remoteName)
		if ctx.Flags.Group {
			groups := core.ConvertToGroups(repos)
			for _, group := range groups {
				fmt.Println(group)
			}
			return nil
		}
		for _, repo := range repos {
			fmt.Println(repo.Name)
		}
		return nil
	},
})

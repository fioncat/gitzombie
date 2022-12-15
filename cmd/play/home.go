package play

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

var Home = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:  "play [name]",
	Desc: "Enter a playground",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(compRepo)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		rootDir := config.Get().Playground
		repo, err := core.SelectLocalRepository(rootDir, ctx.Arg(0))
		if err != nil {
			return err
		}
		fmt.Println(repo.Path)
		return nil
	},
})

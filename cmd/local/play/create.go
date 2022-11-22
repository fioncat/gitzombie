package play

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	Builder string
}

var Create = app.Register(&app.Command[CreateFlags, app.Empty]{
	Use:    "play [-b builder] {play}",
	Desc:   "Create a playground",
	Action: "Create",

	Prepare: func(cmd *cobra.Command, flags *CreateFlags) {
		cmd.Flags().StringVarP(&flags.Builder, "builder", "b", "", "builder name")
		cmd.RegisterFlagCompletionFunc("builder", app.Comp(app.CompBuilder))
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(compGroup)
	},

	Run: func(ctx *app.Context[CreateFlags, app.Empty]) error {
		var err error
		builder := core.DefaultBuilder
		if ctx.Flags.Builder != "" {
			builder, err = core.GetBuilder(ctx.Flags.Builder)
			if err != nil {
				return err
			}
		}
		rootDir := config.Get().Playground
		repo, err := core.NewLocalRepository(rootDir, ctx.Arg(0))
		if err != nil {
			return err
		}
		err = builder.Prepare(nil, repo)
		if err != nil {
			return err
		}

		err = builder.Execute()
		if err != nil {
			return err
		}

		fmt.Println(repo.Path)
		return nil
	},
})

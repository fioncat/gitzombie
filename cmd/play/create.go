package play

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	Builder string
}

var Create = app.Register(&app.Command[CreateFlags, app.Empty]{
	Use:    "play [-b builder] {name}",
	Desc:   "Create a playground",
	Action: "Create",

	Prepare: func(cmd *cobra.Command, flags *CreateFlags) {
		cmd.Flags().StringVarP(&flags.Builder, "builder", "b", "", "builder name")
		cmd.RegisterFlagCompletionFunc("builder", app.Comp(app.CompBuilder))
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(compGroup)
	},

	Run: func(ctx *app.Context[CreateFlags, app.Empty]) error {
		name := ctx.Arg(0)
		rootDir := config.Get().Playground
		repo, err := core.NewLocalRepository(rootDir, name)
		if err != nil {
			return err
		}
		err = repo.Setenv()
		if err != nil {
			return err
		}

		var builder *core.Builder
		if ctx.Flags.Builder != "" {
			builder, err = core.GetBuilder(ctx.Flags.Builder)
			if err != nil {
				return err
			}
		}
		if builder == nil {
			builder = core.DefaultBuilder
		}

		err = builder.Run(repo)
		if err != nil {
			return err
		}

		term.Print("")
		term.Print("playground green|%s| has been created", repo.Name)
		return nil
	},
})

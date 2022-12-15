package builder

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/spf13/cobra"
)

var Delete = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "builder {name}",
	Desc:   "Delete a builder",
	Action: "Delete",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompBuilder)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := ctx.Arg(0)
		path := config.GetDir("builders", name+".yaml")
		return app.Delete(ctx.Arg(0), path)
	},
})

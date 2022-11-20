package delete

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/spf13/cobra"
)

var Job = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "job {name}",
	Desc:   "Delete a job",
	Action: "Delete",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompJob)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := ctx.Arg(0)
		path := config.GetDir("jobs", name+".sh")
		return do(ctx.Arg(0), path)
	},
})

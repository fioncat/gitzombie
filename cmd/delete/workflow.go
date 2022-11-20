package delete

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/spf13/cobra"
)

var Workflow = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "workflow {name}",
	Desc:   "Delete a workflow",
	Action: "Delete",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompWorkflow)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := ctx.Arg(0)
		path := config.GetDir("workflows", name+".yaml")
		return do(ctx.Arg(0), path)
	},
})

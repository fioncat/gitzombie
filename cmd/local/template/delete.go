package template

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

var Delete = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "template {name}",
	Desc:   "Delete a template",
	Action: "Delete",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompTemplate)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := ctx.Arg(0)
		return core.DeleteTemplate(name)
	},
})

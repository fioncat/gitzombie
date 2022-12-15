package delete

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

var Secret = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "secret {key}",
	Desc:   "Delete a secret",
	Action: "Delete",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		key := ctx.Arg(0)
		return core.DeleteSecret(key)
	},
})

package secret

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

var Describe = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "secret {key}",
	Desc:   "Show secret value",
	Action: "Describe",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		key := ctx.Arg(0)
		value, err := core.GetSecret(key, false)
		if err != nil {
			return err
		}
		fmt.Print(value)
		return nil
	},
})

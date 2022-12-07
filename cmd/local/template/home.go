package template

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/spf13/cobra"
)

var Home = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:  "template {name}",
	Desc: "Enter a template",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompTemplate)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := ctx.Arg(0)
		path, err := core.GetTemplate(name)
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
})

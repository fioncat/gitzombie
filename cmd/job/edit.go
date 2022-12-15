package job

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/spf13/cobra"
)

var Edit = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "job {job}",
	Desc:   "Edit job file",
	Action: "Edit",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompJob)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := fmt.Sprintf("%s.sh", ctx.Arg(0))
		path := config.GetDir("jobs", name)
		return app.Edit(path, "", name, nil)
	},
})

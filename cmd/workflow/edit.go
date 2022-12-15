package workflow

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/validate"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Edit = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "workflow {workflow}",
	Desc:   "Edit workflow file",
	Action: "Edit",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompWorkflow)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := fmt.Sprintf("%s.yaml", ctx.Arg(0))
		path := config.GetDir("workflows", name)
		return app.Edit(path, config.DefaultWorkflow, name, func(s string) error {
			data := []byte(s)
			var wf core.Workflow
			err := yaml.Unmarshal(data, &wf)
			if err != nil {
				return errors.Trace(err, "parse yaml")
			}
			return validate.Do(&wf)
		})
	},
})

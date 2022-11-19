package edit

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Builder = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "builder {builder}",
	Desc:   "Edit builder file",
	Action: "Edit",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompBuilder)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := fmt.Sprintf("%s.yaml", ctx.Arg(0))
		path := config.GetDir("builders", name)
		return Do(path, config.DefaultBuilder, name, func(s string) error {
			data := []byte(s)
			var builder core.Builder
			err := yaml.Unmarshal(data, &builder)
			if err != nil {
				return errors.Trace(err, "parse yaml")
			}
			return builder.Validate()
		})
	},
})

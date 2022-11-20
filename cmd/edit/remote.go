package edit

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/validate"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var Remote = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "remote {remote}",
	Desc:   "Edit remote file",
	Action: "Edit",

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote)
	},

	Run: func(ctx *app.Context[app.Empty, app.Empty]) error {
		name := fmt.Sprintf("%s.toml", ctx.Arg(0))
		path := config.GetDir("remotes", name)
		return Do(path, config.DefaultRemote, name, func(s string) error {
			data := []byte(s)
			var remote core.Remote
			err := toml.Unmarshal(data, &remote)
			if err != nil {
				return errors.Trace(err, "parse toml")
			}
			err = validate.Do(&remote)
			return errors.Trace(err, "validate fields")
		})
	},
})

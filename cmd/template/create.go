package template

import (
	"os"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	Path string
}

var Create = app.Register(&app.Command[CreateFlags, app.Empty]{
	Use:    "template [--path path] {name}",
	Desc:   "Create a template",
	Action: "Create",

	Prepare: func(cmd *cobra.Command, flags *CreateFlags) {
		cmd.Flags().StringVarP(&flags.Path, "path", "", "", "target path")
		cmd.Args = cobra.ExactArgs(1)
	},

	Run: func(ctx *app.Context[CreateFlags, app.Empty]) error {
		var err error
		path := ctx.Flags.Path
		if path == "" {
			path, err = os.Getwd()
			if err != nil {
				return errors.Trace(err, "getwd")
			}
		}

		term.PrintOperation("Finding files")
		files, err := core.FindTemplateFiles(path)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return errors.New("no file to write")
		}

		name := ctx.Arg(0)
		return core.CreateTemplate(name, files)
	},
})

package tools

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
)

var Restore = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:  "restore",
	Desc: "Restore all changes",

	RunNoContext: func() error {
		err := git.Exec([]string{"restore", "."}, git.Default)
		if err != nil {
			return err
		}
		return git.Exec([]string{"clean", "-fd"}, git.Default)
	},
})

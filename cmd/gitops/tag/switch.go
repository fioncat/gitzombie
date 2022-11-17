package tag

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
)

var Switch = app.Register(&app.Command[SelectFlag, app.Empty]{
	Use:    "tag [tag]",
	Desc:   "Switch to a tag",
	Action: "Switch",

	Prepare: prepareSelect,

	Run: func(ctx *app.Context[SelectFlag, app.Empty]) error {
		tag, err := doSelect(ctx)
		if err != nil {
			return err
		}
		return git.Checkout(tag, false, git.Default)
	},
})

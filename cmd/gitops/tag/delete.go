package tag

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
)

var Delete = app.Register(&app.Command[SelectFlag, app.Empty]{
	Use:    "tag {tag}",
	Desc:   "Delete a tag",
	Action: "Delete",

	Prepare: prepareSelect,

	Run: func(ctx *app.Context[SelectFlag, app.Empty]) error {
		tag, err := doSelect(ctx)
		if err != nil {
			return err
		}
		term.ConfirmExit("Do you want to delete tag %q", tag)
		err = git.DeleteTag(tag, git.Default)
		if err != nil {
			return err
		}
		return git.PushTag(tag, ctx.Flags.Remote, true, git.Default)
	},
})

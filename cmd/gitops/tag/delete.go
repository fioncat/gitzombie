package tag

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
)

var Delete = app.Register(&app.Command[gitops.RemoteFlags, app.Empty]{
	Use:    "tag {tag}",
	Desc:   "Delete a tag",
	Action: "Delete",

	Prepare: gitops.PreapreRemoteFlags,

	Run: func(ctx *app.Context[gitops.RemoteFlags, app.Empty]) error {
		tag := ctx.Arg(0)
		term.ConfirmExit("Do you want to delete tag %q", tag)
		err := git.DeleteTag(tag, git.Default)
		if err != nil {
			return err
		}
		return git.PushTag(tag, ctx.Flags.Remote, true, git.Default)
	},
})

package tag

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
)

var Create = app.Register(&app.Command[gitops.CreateFlags, app.Empty]{
	Use:  "tag {tag}",
	Desc: "Create a tag and push to remote",

	Prepare: gitops.PrepareCreate,

	Run: func(ctx *app.Context[gitops.CreateFlags, app.Empty]) error {
		tag := ctx.Arg(0)
		err := git.CreateTag(tag, git.Default)
		if err != nil {
			return errors.Trace(err, "create tag")
		}
		if !ctx.Flags.NoPush {
			err = git.PushTag(tag, ctx.Flags.Remote, false, git.Default)
			if err != nil {
				return errors.Trace(err, "push tag")
			}
		}
		return nil
	},
})

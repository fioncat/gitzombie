package tag

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type SwitchFlags struct {
	gitops.RemoteFlags
	gitops.FetchFlags
}

var Switch = app.Register(&app.Command[SwitchFlags, app.Empty]{
	Use:    "tag [tag]",
	Desc:   "Switch to a tag",
	Action: "Switch",

	Prepare: func(cmd *cobra.Command, flags *SwitchFlags) {
		gitops.PreapreRemoteFlags(cmd, &flags.RemoteFlags)
		gitops.PreapreFetchFlags(cmd, &flags.FetchFlags)
		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompGitTag)
	},

	Run: func(ctx *app.Context[SwitchFlags, app.Empty]) error {
		tag := ctx.Arg(0)
		if tag == "" {
			if !ctx.Flags.NoFetch {
				err := git.Fetch(ctx.Flags.Remote, false, true, git.Default)
				if err != nil {
					return err
				}
			}
			tags, err := git.ListTags(git.Default)
			if err != nil {
				return err
			}
			if len(tags) == 0 {
				return errors.New("no tag")
			}
			idx, err := term.FuzzySearch("tag", tags)
			if err != nil {
				return err
			}
			tag = tags[idx]
		}
		return git.Checkout(tag, false, git.Default)
	},
})

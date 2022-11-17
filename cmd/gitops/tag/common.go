package tag

import (
	"errors"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type SelectFlag struct {
	gitops.RemoteFlags
	gitops.FetchFlags
}

func prepareSelect(cmd *cobra.Command, flags *SelectFlag) {
	gitops.PreapreRemoteFlags(cmd, &flags.RemoteFlags)
	gitops.PreapreFetchFlags(cmd, &flags.FetchFlags)
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.ValidArgsFunction = app.Comp(app.CompGitTag)
}

func doSelect(ctx *app.Context[SelectFlag, app.Empty]) (string, error) {
	tag := ctx.Arg(0)
	if tag == "" {
		if !ctx.Flags.NoFetch {
			err := git.Fetch(ctx.Flags.Remote, false, true, git.Default)
			if err != nil {
				return "", err
			}
		}
		tags, err := git.ListTags(git.Default)
		if err != nil {
			return "", err
		}
		if len(tags) == 0 {
			return "", errors.New("no tag")
		}
		idx, err := term.FuzzySearch("tag", tags)
		if err != nil {
			return "", err
		}
		tag = tags[idx]
	}
	return tag, nil
}

package repo

import (
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Open struct{}

func (o *Open) Use() string    { return "repo [remote] [repo]" }
func (o *Open) Desc() string   { return "open repo in default browser" }
func (o *Open) Action() string { return "open" }

func (o *Open) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.MaximumNArgs(2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compRepo)
}

func (o *Open) Run(ctx *Context, args common.Args) error {
	ctx.store.ReadOnly()
	var apiRepo *api.Repository
	var err error
	switch len(args) {
	case 0:
		repo, err := getCurrent(ctx)
		if err != nil {
			return err
		}
		apiRepo, err = apiGet(ctx, repo)

	default:
		apiRepo, err = apiSearch(ctx, args.Get(1))
	}
	if err != nil {
		return err
	}

	return term.Open(apiRepo.WebURL)
}

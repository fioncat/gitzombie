package repo

import (
	"os"

	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Delete struct{}

func (d *Delete) Use() string    { return "repo remote repo" }
func (d *Delete) Desc() string   { return "delete a repo" }
func (d *Delete) Action() string { return "delete" }

func (d *Delete) Prepare(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(2)
	cmd.ValidArgsFunction = common.Comp(compRemote, compRepo)
}

func (d *Delete) Run(ctx *Context, args common.Args) error {
	repo, err := getLocal(ctx, args.Get(1))
	if err != nil {
		return err
	}

	if !term.Confirm("delete %s", repo.Path) {
		return nil
	}
	_, err = os.Stat(repo.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		err = os.RemoveAll(repo.Path)
		if err != nil {
			return errors.Trace(err, "remove repo")
		}
	}

	ctx.store.Delete(repo)
	return nil
}

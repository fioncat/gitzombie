package repo

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Detach struct{}

func (d *Detach) Use() string    { return "detach remote repo" }
func (d *Detach) Desc() string   { return "detach current path" }
func (d *Detach) Action() string { return "" }

func (d *Detach) Prepare(cmd *cobra.Command) {}

func (d *Detach) Run(ctx *Context, args common.Args) error {
	dir, err := git.EnsureCurrent()
	if err != nil {
		return err
	}

	repo, err := ctx.store.GetByPath(dir)
	if err != nil {
		return err
	}

	term.ConfirmExit("Do you want to detach this path from %s", repo.FullName())

	ctx.store.Delete(repo)
	return nil
}

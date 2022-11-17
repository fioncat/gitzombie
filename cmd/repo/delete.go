package repo

import (
	"os"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var Delete = app.Register(&app.Command[app.Empty, Data]{
	Use:    "repo {remote} {repo}",
	Desc:   "delete a repo",
	Action: "Delete",

	Init: initData[app.Empty],

	PrepareNoFlag: func(cmd *cobra.Command) {
		cmd.Args = cobra.ExactArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompRepo)
	},

	Run: func(ctx *app.Context[app.Empty, Data]) error {
		repo, err := getLocal(ctx, ctx.Arg(1))
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

		ctx.Data.Store.Delete(repo)
		return nil
	},
})

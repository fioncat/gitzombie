package repo

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
)

var Detach = app.Register(&app.Command[app.Empty, Data]{
	Use:  "detach {remote} {repo}",
	Desc: "Detach current path",

	Init: initData[app.Empty],

	Run: func(ctx *app.Context[app.Empty, Data]) error {
		dir, err := git.EnsureCurrent()
		if err != nil {
			return err
		}

		repo, err := ctx.Data.Store.GetByPath(dir)
		if err != nil {
			return err
		}

		term.ConfirmExit("Do you want to detach this path from %s", repo.FullName())

		ctx.Data.Store.Delete(repo)
		return nil
	},
})

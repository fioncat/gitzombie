package repo

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
)

var Detach = app.Register(&app.Command[struct{}, Data]{
	Use:  "detach {remote} {repo}",
	Desc: "detach current path",

	Init: initData[struct{}],

	Run: func(ctx *app.Context[struct{}, Data]) error {
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

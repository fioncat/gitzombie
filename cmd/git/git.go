package git

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/pkg/git"
)

type Context struct {
	Path string
}

type App struct{}

func (app *App) BuildContext(args common.Args) (*Context, error) {
	ctx := new(Context)
	dir, err := git.EnsureCurrent()
	if err != nil {
		return nil, err
	}
	ctx.Path = dir
	return ctx, nil
}

func (app *App) Ops() []common.Operation[Context] {
	return []common.Operation[Context]{
		&SyncBranch{}, &CreateBranch{},
	}
}

func (app *App) Close(ctx *Context) error { return nil }

package repo

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
)

type Context struct {
	store *core.RepositoryStorage

	remote *core.Remote
}

type App struct{}

func (app *App) BuildContext(args common.Args) (*Context, error) {
	ctx := new(Context)
	var err error
	ctx.store, err = core.NewRepositoryStorage()
	if err != nil {
		return nil, err
	}
	if args.Get(0) != "" {
		remote, err := getRemote(args.Get(0))
		if err != nil {
			return nil, err
		}
		ctx.remote = remote
	}
	return ctx, nil
}

func (app *App) Ops() []common.Operation[Context] {
	return []common.Operation[Context]{
		&Home{},
		&List{},
		&Attach{},
		&Delete{},
		&Detach{},
		&Open{},
		&Merge{},
	}
}

func (app *App) Close(ctx *Context) error {
	err := ctx.store.Close()
	if err != nil {
		return errors.Trace(err, "save data")
	}
	return nil
}

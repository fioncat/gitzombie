package branch

import "github.com/fioncat/gitzombie/cmd/common"

type App struct {
	common.NoneContext
}

func (app *App) Ops() []common.Operation[struct{}] {
	return []common.Operation[struct{}]{
		&Create{}, &Sync{},
	}
}

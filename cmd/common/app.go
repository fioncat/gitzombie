package common

import (
	"fmt"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	ErrNotGit = errors.New("you are not int a git repository")
)

type App[Context any] interface {
	BuildContext(args Args) (*Context, error)
	Ops() []Operation[Context]

	Close(ctx *Context) error
}

type Operation[Context any] interface {
	Use() string
	Desc() string
	Action() string

	Prepare(cmd *cobra.Command)
	Run(ctx *Context, args Args) error
}

var actionMap map[string]*cobra.Command

func Build[Context any](cmd *cobra.Command, app App[Context]) {
	ops := app.Ops()
	for _, op := range ops {
		op := op
		appCmd := &cobra.Command{
			Use:   op.Use(),
			Short: op.Desc(),

			RunE: func(_ *cobra.Command, args []string) error {
				cargs := Args(args)
				ctx, err := app.BuildContext(cargs)
				if err != nil {
					return err
				}
				err = op.Run(ctx, cargs)
				if err != nil {
					return err
				}
				return app.Close(ctx)
			},
		}
		op.Prepare(appCmd)

		if action := op.Action(); action != "" {
			if actionMap == nil {
				actionMap = make(map[string]*cobra.Command)
			}
			actionCmd := actionMap[action]
			if actionCmd == nil {
				actionCmd = &cobra.Command{
					Use:   action,
					Short: fmt.Sprintf("%s actions", action),
				}
				actionMap[action] = actionCmd
				cmd.AddCommand(actionCmd)
			}
			actionCmd.AddCommand(appCmd)
		} else {
			cmd.AddCommand(appCmd)
		}
	}
}

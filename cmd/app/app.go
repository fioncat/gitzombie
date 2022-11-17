package app

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type Context[Flags, Data any] struct {
	args []string

	Flags *Flags
	Data  *Data

	closeFuncs []func() error
}

func (ctx *Context[Flags, Data]) ArgDefault(idx int, def string) string {
	if idx < 0 || idx >= len(ctx.args) {
		return def
	}
	return ctx.args[idx]
}

func (ctx *Context[Flags, Data]) Arg(idx int) string {
	return ctx.ArgDefault(idx, "")
}

func (ctx *Context[Flags, Data]) ArgLen() int {
	return len(ctx.args)
}

func (ctx *Context[Flags, Data]) OnClose(f func() error) {
	ctx.closeFuncs = append(ctx.closeFuncs, f)
}

type PrepareFunction[Flags any] func(cmd *cobra.Command, flags *Flags)

type InitFunction[Flags, Data any] func(ctx *Context[Flags, Data]) error

type RunFunction[Flags, Data any] func(ctx *Context[Flags, Data]) error

type Command[Flags, Data any] struct {
	Use    string
	Desc   string
	Action string

	Prepare PrepareFunction[Flags]
	Init    InitFunction[Flags, Data]
	Run     RunFunction[Flags, Data]
}

var actions = map[string]*cobra.Command{}

func Register[Flags, Data any](app *Command[Flags, Data]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   app.Use,
		Short: app.Desc,
	}
	var flags *Flags
	if app.Prepare != nil {
		flags = new(Flags)
		app.Prepare(cmd, flags)
	}
	cmd.RunE = func(_ *cobra.Command, args []string) error {
		ctx := &Context[Flags, Data]{
			args:  args,
			Flags: flags,
		}
		if app.Init != nil {
			err := app.Init(ctx)
			if err != nil {
				return err
			}
		}
		if app.Run != nil {
			err := app.Run(ctx)
			if err != nil {
				return err
			}
		}
		for _, closeFunc := range ctx.closeFuncs {
			err := closeFunc()
			if err != nil {
				return err
			}
		}
		return nil
	}
	if app.Action == "" {
		name := strings.Split(app.Use, " ")[0]
		actions[name] = cmd
		return cmd
	}

	actionCmd := actions[app.Action]
	if actionCmd == nil {
		actionCmd = &cobra.Command{
			Use:   strings.ToLower(app.Action),
			Short: fmt.Sprintf("%s actions", app.Action),
		}
		actions[app.Action] = actionCmd
	}
	actionCmd.AddCommand(cmd)
	return cmd
}

func Root(cmd *cobra.Command) {
	for _, action := range actions {
		cmd.AddCommand(action)
	}
}

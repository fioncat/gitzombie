package cmd

import "github.com/spf13/cobra"

type App[Context any] interface {
	Name() string

	BuildContext(args []string) (*Context, error)
	Ops() []Operation[Context]

	Close(ctx *Context) error
}

type Operation[Context any] interface {
	Name() []string
	Desc() string

	Prepare(cmd *cobra.Command)
	Handle(ctx *Context) error
}

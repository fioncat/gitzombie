package run

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type WorkflowFlags struct {
	Yes bool

	Current bool
}

var Workflow = app.Register(&app.Command[WorkflowFlags, app.Empty]{
	Use:    "workflow [-y] [-c] {workflow}",
	Desc:   "Run workflow",
	Action: "Run",

	Prepare: func(cmd *cobra.Command, flags *WorkflowFlags) {
		cmd.Flags().BoolVarP(&flags.Yes, "yes", "y", false, "donot confirm")
		cmd.Flags().BoolVarP(&flags.Current, "current", "c", false, "run workflow on current repo")

		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompWorkflow)
	},

	Run: func(ctx *app.Context[WorkflowFlags, app.Empty]) error {
		wf, err := core.GetWorkflow(ctx.Arg(0))
		if err != nil {
			return err
		}
		// TODO: handle current

		if wf.Select == nil {
			term.Print("nothing to do")
			return nil
		}

		store, err := core.NewRepositoryStorage()
		if err != nil {
			return err
		}
		store.ReadOnly()
		items, err := wf.Select.Match(store)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			term.Print("no repo selected, please check your select")
			return nil
		}
		for _, item := range items {
			fmt.Println(item.Path)
		}

		return nil
	},
})

package cmd

import (
	"fmt"
	"runtime"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/fioncat/gitzombie/pkg/worker"
	"github.com/fioncat/gitzombie/scripts"
	"github.com/spf13/cobra"

	_ "github.com/fioncat/gitzombie/cmd/delete"
	_ "github.com/fioncat/gitzombie/cmd/edit"
	_ "github.com/fioncat/gitzombie/cmd/gitops/branch"
	_ "github.com/fioncat/gitzombie/cmd/gitops/tag"
	_ "github.com/fioncat/gitzombie/cmd/local/play"
	_ "github.com/fioncat/gitzombie/cmd/repo"
	_ "github.com/fioncat/gitzombie/cmd/run"
)

var Root = &cobra.Command{
	Use: "gitzombie",

	SilenceErrors: true,
	SilenceUsage:  true,

	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		err := config.Init()
		return errors.Trace(err, "init")
	},

	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Print init scripts",
}

var initZsh = &cobra.Command{
	Use:   "zsh",
	Short: "Print init scripts, you can add `source <(gitzombie init zsh)` to your profile",

	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println(scripts.Common)
		fmt.Println(scripts.ZshComp)
		return nil
	},
}

var initBash = &cobra.Command{
	Use:   "bash",
	Short: "Print init scripts, you can add `source <(gitzombie init bash)` to your profile",

	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println(scripts.Common)
		fmt.Println(scripts.BashComp)
		return nil
	},
}

func init() {
	initCmd.AddCommand(initZsh, initBash)
	Root.AddCommand(initCmd)
	app.Root(Root)

	Root.PersistentFlags().BoolVarP(&term.AlwaysYes, "yes", "", false, "donot confirm")
	Root.PersistentFlags().IntVarP(&worker.Count, "worker", "", runtime.NumCPU(), "number of workers")
}

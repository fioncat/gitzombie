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

	_ "github.com/fioncat/gitzombie/api/github"
	_ "github.com/fioncat/gitzombie/api/gitlab"

	_ "github.com/fioncat/gitzombie/cmd/delete"
	_ "github.com/fioncat/gitzombie/cmd/edit"
	_ "github.com/fioncat/gitzombie/cmd/gitops/branch"
	_ "github.com/fioncat/gitzombie/cmd/gitops/tag"
	_ "github.com/fioncat/gitzombie/cmd/gitops/tools"
	_ "github.com/fioncat/gitzombie/cmd/local/play"
	_ "github.com/fioncat/gitzombie/cmd/local/template"
	_ "github.com/fioncat/gitzombie/cmd/repo"
	_ "github.com/fioncat/gitzombie/cmd/run"
	_ "github.com/fioncat/gitzombie/cmd/secret"
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

var noAlias bool

func addInitCommand(shell string) {
	cmd := &cobra.Command{
		Use:   shell,
		Short: fmt.Sprintf("Print init scripts, you can add `source <(gitzombie init %s)` to your profile", shell),

		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(scripts.Common)
			fmt.Println(scripts.Comps[shell])
			if !noAlias {
				fmt.Println(scripts.Alias)
			}
		},
	}
	cmd.Flags().BoolVarP(&noAlias, "no-alias", "", false, "disable alias")
	initCmd.AddCommand(cmd)
}

func init() {
	addInitCommand("zsh")
	addInitCommand("bash")
	Root.AddCommand(initCmd)

	app.Root(Root)

	Root.PersistentFlags().BoolVarP(&term.AlwaysYes, "yes", "", false, "donot confirm")
	Root.PersistentFlags().IntVarP(&worker.Count, "worker", "", runtime.NumCPU(), "number of workers")
}

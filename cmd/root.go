package cmd

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/cmd/gitops/branch"
	"github.com/fioncat/gitzombie/cmd/repo"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/scripts"
	"github.com/spf13/cobra"
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
		fmt.Println(scripts.ZshComp)
		fmt.Println(scripts.Home)
		return nil
	},
}

var initBash = &cobra.Command{
	Use:   "bash",
	Short: "Print init scripts, you can add `source <(gitzombie init bash)` to your profile",

	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println(scripts.BashComp)
		fmt.Println(scripts.Home)
		return nil
	},
}

func init() {
	common.Build[repo.Context](Root, &repo.App{})
	common.Build[struct{}](Root, &branch.App{})

	initCmd.AddCommand(initZsh, initBash)
	Root.AddCommand(initCmd)
}

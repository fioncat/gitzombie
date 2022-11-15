package cmd

import (
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/cmd/repo"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
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

func init() {
	common.Build[repo.Context](Root, &repo.App{})
}

package app

import (
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

const CompNoSpaceFlag = cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp

type CompResult struct {
	Items []string
	Flag  cobra.ShellCompDirective
}

var EmptyCompResult = &CompResult{
	Flag: cobra.ShellCompDirectiveNoFileComp,
}

type CompAction func(args []string) (*CompResult, error)

func Comp(actions ...CompAction) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(actions) == 0 {
		return nil
	}
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		err := config.Init()
		if err != nil {
			term.Warn("complete: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}
		idx := len(args)
		if idx >= len(actions) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		action := actions[idx]
		result, err := action(args)
		if err != nil {
			term.Warn("complete: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}
		if result.Flag <= 0 {
			result.Flag = cobra.ShellCompDirectiveNoFileComp
		}
		return result.Items, result.Flag
	}
}

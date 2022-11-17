package gitops

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	RemoteFlags
	NoPush bool
}

func PrepareCreate(cmd *cobra.Command, flags *CreateFlags) {
	PreapreRemoteFlags(cmd, &flags.RemoteFlags)
	cmd.Flags().BoolVarP(&flags.NoPush, "no-push", "", false, "donot push to remote")
	cmd.Args = cobra.ExactArgs(1)
}

type RemoteFlags struct {
	Remote string
}

func PreapreRemoteFlags(cmd *cobra.Command, flags *RemoteFlags) {
	cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "origin", "remote name")
	cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompGitRemote))
}

type FetchFlags struct {
	NoFetch bool
}

func PreapreFetchFlags(cmd *cobra.Command, flags *FetchFlags) {
	cmd.Flags().BoolVarP(&flags.NoFetch, "no-fetch", "", false, "no fetch remote")
}

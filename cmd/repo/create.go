package repo

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type CreateFlags struct {
	Builder string
}

var Create = app.Register(&app.Command[CreateFlags, Data]{
	Use:    "repo [-b builder] {remote} {name}",
	Desc:   "Create a repo",
	Action: "Create",

	Init: initData[CreateFlags],

	Prepare: func(cmd *cobra.Command, flags *CreateFlags) {
		cmd.Flags().StringVarP(&flags.Builder, "builder", "b", "", "builder name")
		cmd.RegisterFlagCompletionFunc("builder", app.Comp(app.CompBuilder))
		cmd.Args = cobra.ExactArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompGroup)
	},

	Run: func(ctx *app.Context[CreateFlags, Data]) error {
		repo, err := core.CreateRepository(ctx.Data.Remote, ctx.Arg(1))
		if err != nil {
			return err
		}

		term.ConfirmExit("create %s", repo.Path)

		localRepo := repo.ToLocal()
		err = localRepo.Setenv()
		if err != nil {
			return err
		}
		user, email := ctx.Data.Remote.GetUserEmail(repo)
		url, err := ctx.Data.Remote.GetCloneURL(repo)
		if err != nil {
			return err
		}

		err = osutil.Setenv(map[string]string{
			"REPO_USER":  user,
			"REPO_EMAIL": email,
			"REPO_URL":   url,
		})
		if err != nil {
			return err
		}
		err = ctx.Data.Remote.Setenv()
		if err != nil {
			return err
		}

		var builder *core.Builder
		if ctx.Flags.Builder != "" {
			builder, err = core.GetBuilder(ctx.Flags.Builder)
			if err != nil {
				return err
			}
		}
		if builder == nil {
			builder = core.DefaultBuilder
		}

		err = builder.Run(localRepo)
		if err != nil {
			return err
		}

		err = ctx.Data.Store.Add(repo)
		if err != nil {
			return err
		}

		term.Print("")
		term.Print("repo green|%s| has been created", repo.FullName())
		return nil
	},
})

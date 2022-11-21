package repo

import (
	"fmt"
	"strings"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/spf13/cobra"
)

type JumpFlags struct {
	Remote string
}

type JumpData struct {
	RepoStore    *core.RepositoryStorage
	KeywordStore *core.JumpKeywordStorage

	Remotes []string
}

var Jump = app.Register(&app.Command[JumpFlags, JumpData]{
	Use:  "jump [-r remote] [keyword]",
	Desc: "Auto jump to a repo",

	Prepare: func(cmd *cobra.Command, flags *JumpFlags) {
		cmd.Args = cobra.MaximumNArgs(1)
		cmd.ValidArgsFunction = app.Comp(compJump)

		cmd.Flags().StringVarP(&flags.Remote, "remote", "r", "", "remote name")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompRemote))
	},

	Init: func(ctx *app.Context[JumpFlags, JumpData]) error {
		repoStore, err := core.NewRepositoryStorage()
		if err != nil {
			return err
		}
		ctx.OnClose(func() error {
			err = repoStore.Close()
			return errors.Trace(err, "close repo storage")
		})

		jumpStore, err := core.NewJumpKeywordStorage()
		if err != nil {
			return err
		}
		ctx.OnClose(func() error {
			err = jumpStore.Close()
			return errors.Trace(err, "close jump keyword storage")
		})

		var remotes []string
		if ctx.Flags.Remote != "" {
			remotes = []string{ctx.Flags.Remote}
		} else {
			remotes, err = core.ListRemoteNames()
			if err != nil {
				return errors.Trace(err, "list remotes")
			}
		}

		ctx.Data = &JumpData{
			RepoStore:    repoStore,
			KeywordStore: jumpStore,
			Remotes:      remotes,
		}
		return nil
	},

	Run: func(ctx *app.Context[JumpFlags, JumpData]) error {
		var repos []*core.Repository
		for _, remote := range ctx.Data.Remotes {
			remoteRepos := ctx.Data.RepoStore.List(remote)
			repos = append(repos, remoteRepos...)
		}
		if len(repos) == 0 {
			return errors.New("no repo")
		}
		core.SortRepositories(repos)

		repo := jumpSelectRepo(repos, ctx)
		if repo == nil {
			return errors.New("cannot find match repo")
		}

		repo.MarkAccess()
		fmt.Println(repo.Path)
		return nil
	},
})

func jumpSelectRepo(repos []*core.Repository, ctx *app.Context[JumpFlags, JumpData]) *core.Repository {
	query := ctx.Arg(0)
	if query == "" {
		return repos[0]
	}

	for _, repo := range repos {
		if strings.Contains(repo.Name, query) {
			ctx.Data.KeywordStore.Add(query)
			return repo
		}
	}
	return nil
}

func compJump(_ []string) (*app.CompResult, error) {
	store, err := core.NewJumpKeywordStorage()
	if err != nil {
		return nil, err
	}
	kws := store.List()
	err = store.Close()
	if err != nil {
		return nil, errors.Trace(err, "close keyword storage")
	}
	return &app.CompResult{Items: kws}, nil
}

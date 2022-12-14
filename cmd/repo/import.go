package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/fioncat/gitzombie/pkg/worker"
	"github.com/spf13/cobra"
)

type ImportFlags struct {
	Ignore  []string
	LogPath string
}

var Import = app.Register(&app.Command[ImportFlags, core.RepositoryStorage]{
	Use:  "import [-i ignore-repo]... {remote} {group}",
	Desc: "Import repos to workspace",

	Init: initData[ImportFlags],

	Prepare: func(cmd *cobra.Command, flags *ImportFlags) {
		cmd.Flags().StringSliceVarP(&flags.Ignore, "ignore", "i", nil, "ignore repo pattern")
		cmd.Flags().StringVarP(&flags.LogPath, "log-path", "", "", "log path")

		cmd.Args = cobra.ExactArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompGroup)
	},

	Run: func(ctx *app.Context[ImportFlags, core.RepositoryStorage]) error {
		group := ctx.Arg(1)
		group = strings.Trim(group, "/")

		remote, err := core.GetRemote(ctx.Arg(0))
		if err != nil {
			return err
		}

		var apiRepos []*api.Repository
		err = api.Exec("list repos", remote, func(p api.Provider) error {
			apiRepos, err = p.ListRepositories(group)
			return err
		})
		if err != nil {
			return err
		}
		if len(ctx.Flags.Ignore) > 0 {
			apiRepos, err = importFilterIgnore(apiRepos, ctx.Flags.Ignore)
			if err != nil {
				return err
			}
		}
		if len(apiRepos) == 0 {
			term.Println("no repo to import")
			return nil
		}
		apiRepos, err = term.EditItems(config.Get().Editor, apiRepos,
			func(repo *api.Repository) string {
				return repo.Name
			})
		if err != nil {
			return err
		}
		tasks, err := importGetTasks(ctx, remote, apiRepos)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			if len(apiRepos) <= 1 {
				term.Println("nothing to do, repo is already exists")
			} else {
				term.Println("nothing to do, all repos are already exists")
			}
			return nil
		}
		repoWord := english.Plural(len(tasks), "repo", "repos")
		term.ConfirmExit("Do you want to clone %s", repoWord)
		w := worker.Worker[CloneTask]{
			Name: "import",

			Tasks:   tasks,
			Tracker: worker.NewJobTracker[CloneTask]("cloning"),

			LogPath: ctx.Flags.LogPath,
		}
		return w.Run(func(task *worker.Task[CloneTask]) error {
			return task.Value.Execute()
		})
	},
})

func importFilterIgnore(apiRepos []*api.Repository, ignores []string) ([]*api.Repository, error) {
	newRepos := make([]*api.Repository, 0, len(apiRepos))
	for _, apiRepo := range apiRepos {
		var shouldIgnore bool
		for _, ignore := range ignores {
			match, err := filepath.Match(ignore, apiRepo.Name)
			if err != nil {
				return nil, fmt.Errorf("invalid ignore %q: %v", ignore, err)
			}
			if match {
				shouldIgnore = true
				break
			}
		}
		if !shouldIgnore {
			newRepos = append(newRepos, apiRepo)
		}
	}
	return newRepos, nil
}

func importGetTasks(ctx *app.Context[ImportFlags, core.RepositoryStorage], remote *core.Remote, apiRepos []*api.Repository) ([]*worker.Task[CloneTask], error) {
	var err error
	tasks := make([]*worker.Task[CloneTask], 0, len(apiRepos))
	for _, apiRepo := range apiRepos {
		repo := ctx.Data.GetByName(remote.Name, apiRepo.Name)
		if repo == nil {
			repo, err = core.WorkspaceRepository(remote, apiRepo.Name)
			if err != nil {
				return nil, errors.Trace(err, "convert repo %q", apiRepo.Name)
			}
			err = ctx.Data.Add(repo)
			if err != nil {
				return nil, errors.Trace(err, "add repo %s", repo.Name)
			}
		}
		exists, err := osutil.DirExists(repo.Path)
		if err != nil {
			return nil, errors.Trace(err, "check repo exists")
		}
		if exists {
			continue
		}
		url, err := remote.GetCloneURL(repo)
		if err != nil {
			return nil, errors.Trace(err, "get clone url")
		}
		user, email := remote.GetUserEmail(repo)
		tasks = append(tasks, &worker.Task[CloneTask]{
			Name: repo.Name,
			Value: &CloneTask{
				Path:  repo.Path,
				URL:   url,
				User:  user,
				Email: email,
			},
		})
	}
	return tasks, nil
}

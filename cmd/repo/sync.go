package repo

import (
	"fmt"
	"path/filepath"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/fioncat/gitzombie/pkg/worker"
)

type SyncFlags struct {
	LogPath string
}

type SyncData struct {
	Store *core.RepositoryStorage

	Remotes   []*core.Remote
	remoteMap map[string]*core.Remote

	Repos []*core.Repository

	WorkspaceRepos []*core.Repository
}

var Sync = app.Register(&app.Command[SyncFlags, SyncData]{
	Use:    "repo",
	Desc:   "Sync workspace repo",
	Action: "Sync",

	Init: func(ctx *app.Context[SyncFlags, SyncData]) error {
		store, err := core.NewRepositoryStorage()
		if err != nil {
			return errors.Trace(err, "init repo storage")
		}
		remoteNames, err := core.ListRemoteNames()
		if err != nil {
			return errors.Trace(err, "list remotes")
		}
		remotes := make([]*core.Remote, len(remoteNames))
		remoteMap := make(map[string]*core.Remote, len(remoteNames))
		for i, name := range remoteNames {
			remote, err := core.GetRemote(name)
			if err != nil {
				return errors.Trace(err, "get remote %q", name)
			}
			remotes[i] = remote
			remoteMap[name] = remote
		}

		var repos []*core.Repository
		for _, remote := range remotes {
			remoteRepos := store.List(remote.Name)
			repos = append(repos, remoteRepos...)
		}

		var wpRepos []*core.Repository
		for _, remote := range remotes {
			rootDir := filepath.Join(config.Get().Workspace, remote.Name)
			exists, err := osutil.DirExists(rootDir)
			if err != nil {
				return errors.Trace(err, "check remote dir exists")
			}
			if !exists {
				continue
			}
			remoteWpRepos, err := core.DiscoverLocalRepositories(rootDir)
			if err != nil {
				return errors.Trace(err, "discover remote repos")
			}
			for _, wpRepo := range remoteWpRepos {
				wpRepo.Remote = remote.Name
			}
			wpRepos = append(wpRepos, remoteWpRepos...)
		}
		ctx.Data = &SyncData{
			Store:          store,
			Remotes:        remotes,
			remoteMap:      remoteMap,
			Repos:          repos,
			WorkspaceRepos: wpRepos,
		}
		ctx.OnClose(func() error { return store.Close() })
		return nil
	},

	Run: func(ctx *app.Context[SyncFlags, SyncData]) error {
		err := syncStorage(ctx)
		if err != nil {
			return err
		}
		return syncWorkspace(ctx)
	},
})

func syncStorage(ctx *app.Context[SyncFlags, SyncData]) error {
	tasks, err := syncBuildCloneTasks(ctx.Data)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}

	repoWord := english.Plural(len(tasks), "repo", "repos")
	ok := term.Confirm("do you want to clone %s", repoWord)
	if !ok {
		term.Print("skip cloning")
		return nil
	}

	w := worker.Worker[CloneTask]{
		Name: "sync",

		Tasks:   tasks,
		Tracker: worker.NewJobTracker[CloneTask]("cloning"),

		LogPath: ctx.Flags.LogPath,
	}
	return w.Run(func(task *worker.Task[CloneTask]) error {
		return task.Value.Execute()
	})
}

func syncBuildCloneTasks(data *SyncData) ([]*worker.Task[CloneTask], error) {
	var tasks []*worker.Task[CloneTask]
	for _, repo := range data.Repos {
		path := repo.Path
		exists, err := osutil.DirExists(path)
		if err != nil {
			return nil, errors.Trace(err, "check repo exists")
		}
		if !exists {
			remote := data.remoteMap[repo.Remote]
			if remote == nil {
				return nil, fmt.Errorf("cannot find remote %q on repo %q", repo.Remote, repo.Name)
			}

			url, err := remote.GetCloneURL(repo)
			if err != nil {
				return nil, errors.Trace(err, "get clone url")
			}
			user, email := remote.GetUserEmail(repo)
			tasks = append(tasks, &worker.Task[CloneTask]{
				Name: repo.FullName(),
				Value: &CloneTask{
					Path:  path,
					URL:   url,
					User:  user,
					Email: email,
				},
			})
		}
	}
	return tasks, nil
}

func syncWorkspace(ctx *app.Context[SyncFlags, SyncData]) error {
	for _, workspaceRepo := range ctx.Data.WorkspaceRepos {
		name := workspaceRepo.Name
		remoteName := workspaceRepo.Remote
		if remoteName == "" {
			continue
		}
		if ctx.Data.Store.GetByName(remoteName, name) == nil {
			remote := ctx.Data.remoteMap[remoteName]
			if remote == nil {
				return fmt.Errorf("cannot find remote %q on repo %q", remoteName, name)
			}

			repo, err := core.WorkspaceRepository(remote, name)
			if err != nil {
				return err
			}
			err = ctx.Data.Store.Add(repo)
			if err != nil {
				return errors.Trace(err, "add repo")
			}
		}
	}
	return nil
}

package repo

import (
	"fmt"
	"path/filepath"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/fioncat/gitzombie/pkg/worker"
	"github.com/spf13/cobra"
)

type DownloadFlags struct {
	Search  bool
	Pattern string
	Output  string

	Tag string

	NoBar bool

	LogPath string

	Latest bool
}

var Download = app.Register(&app.Command[DownloadFlags, Data]{
	Use:  "download {remote} {repo} [-t tag] [-s] [-p pattern] [-o out] [--no-bar]",
	Desc: "Download files from release",

	Init: initData[DownloadFlags],

	Prepare: func(cmd *cobra.Command, flags *DownloadFlags) {
		cmd.Args = cobra.MaximumNArgs(2)
		cmd.ValidArgsFunction = app.Comp(app.CompRemote, app.CompRepo)

		cmd.Flags().BoolVarP(&flags.Search, "search", "s", false, "search repo")
		cmd.Flags().StringVarP(&flags.Pattern, "pattern", "p", "", "file pattern")
		cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "output dir")

		cmd.Flags().StringVarP(&flags.Tag, "tag", "t", "", "tag name")
		cmd.RegisterFlagCompletionFunc("tag", app.Comp(app.CompGitTag))

		cmd.Flags().BoolVarP(&flags.NoBar, "no-bar", "n", false, "donot show download bar")

		cmd.Flags().StringVarP(&flags.LogPath, "log-path", "", "", "error log path")

		cmd.Flags().BoolVarP(&flags.Latest, "latest", "l", false, "use latest release")
	},

	Run: func(ctx *app.Context[DownloadFlags, Data]) error {
		var repo *core.Repository
		var err error
		switch ctx.ArgLen() {
		case 0:
			repo, err = getCurrent(ctx)

		default:
			if ctx.Flags.Search {
				apiRepo, err := apiSearch(ctx, ctx.Arg(1))
				if err != nil {
					return err
				}
				repo, err = core.WorkspaceRepository(ctx.Data.Remote, apiRepo.Name)
				if err != nil {
					return err
				}
			} else {
				repo, err = getLocal(ctx, ctx.Arg(1))
			}
		}
		if err != nil {
			return err
		}

		releases, err := downloadGetRelease(ctx, repo)
		if err != nil {
			return err
		}
		if len(releases) == 0 {
			return errors.New("no release found")
		}
		var release *api.Release
		if len(releases) == 1 {
			release = releases[0]
		} else {
			release, err = downloadSelectRelease(releases)
			if err != nil {
				return err
			}
		}
		files, err := downloadSelectFiles(release, ctx.Flags.Pattern)
		if err != nil {
			return err
		}

		fileWord := english.Plural(len(files), "release file", "")
		term.ConfirmExit("Do you want to download %s", fileWord)

		tasks := make([]*worker.Task[worker.BytesTask], len(files))
		err = execProvider("open release download stream", ctx.Data.Remote, func(p api.Provider) error {
			for i, file := range files {
				reader, err := p.DownloadReleaseFile(repo, file)
				if err != nil {
					return errors.Trace(err, "open file %q", file.Name)
				}
				tasks[i] = worker.DownloadTask(file.Name, reader, uint64(file.Size))
			}
			return nil
		})
		if err != nil {
			return err
		}

		var tracker worker.Tracker[worker.BytesTask]
		if ctx.Flags.NoBar {
			tracker = worker.NewBytesTracker(tasks)
		} else {
			tracker = worker.NewBytesBarTracker(tasks)
		}

		w := worker.Bytes{
			Tracker: tracker,
			Tasks:   tasks,
			LogPath: ctx.Flags.LogPath,
		}
		term.PrintOperation("begin to download %s", fileWord)
		return w.Download(ctx.Flags.Output)
	},
})

func downloadGetRelease(ctx *app.Context[DownloadFlags, Data], repo *core.Repository) ([]*api.Release, error) {
	var err error
	if ctx.Flags.Tag != "" || ctx.Flags.Latest {
		var release *api.Release
		err = execProvider("get release by tag", ctx.Data.Remote, func(p api.Provider) error {
			release, err = p.GetRelease(repo, ctx.Flags.Tag)
			return err
		})
		return []*api.Release{release}, err
	}

	var releases []*api.Release
	err = execProvider("list releases", ctx.Data.Remote, func(p api.Provider) error {
		releases, err = p.ListReleases(repo)
		return err
	})
	return releases, err
}

func downloadSelectRelease(releases []*api.Release) (*api.Release, error) {
	items := make([]string, len(releases))
	for i, release := range releases {
		if release.Tag == release.Name {
			items[i] = release.Name
			continue
		}
		items[i] = fmt.Sprintf("%s (%s)", release.Name, release.Tag)
	}
	idx, err := term.FuzzySearch("release", items)
	if err != nil {
		return nil, err
	}
	return releases[idx], nil
}

func downloadSelectFiles(release *api.Release, pattern string) ([]*api.ReleaseFile, error) {
	if len(release.Files) == 0 {
		return nil, fmt.Errorf("no file in release %s", release.Name)
	}
	files := release.Files
	if pattern != "" {
		files := make([]*api.ReleaseFile, 0, len(release.Files))
		for _, file := range release.Files {
			match, err := filepath.Match(pattern, file.Name)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %q: %v", pattern, err)
			}
			if match {
				files = append(files, file)
			}
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("no file matches pattern %q", pattern)
		}
	}

	files, err := term.EditItems(config.Get().Editor, files, func(file *api.ReleaseFile) string {
		return file.Name
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no file to download after edit")
	}
	return files, nil
}

package repo

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type CleanFlags struct {
	Days   int
	Edit   bool
	Remote []string
	Never  bool
}

type CleanData struct {
	Items []*CleanItem
	Store *core.RepositoryStorage

	RepoPaths []string
}

type CleanItem struct {
	Repo *core.Repository

	Remote *core.Remote

	Days  int
	Never bool
}

var Clean = app.Register(&app.Command[CleanFlags, CleanData]{
	Use:  "clean [--days days] [-e] [-r remote]...",
	Desc: "Clean repos",

	Prepare: func(cmd *cobra.Command, flags *CleanFlags) {
		cmd.Flags().IntVarP(&flags.Days, "days", "d", 30, "days threshold")
		cmd.Flags().BoolVarP(&flags.Never, "never", "n", false, "only delete repo that never access")
		cmd.Flags().BoolVarP(&flags.Edit, "edit", "e", false, "edit items")
		cmd.Flags().StringSliceVarP(&flags.Remote, "remote", "r", nil, "scan remotes")
		cmd.RegisterFlagCompletionFunc("remote", app.Comp(app.CompRemote))
	},

	Init: func(ctx *app.Context[CleanFlags, CleanData]) error {
		if ctx.Flags.Days <= 1 {
			return fmt.Errorf("invalid flag days %d: should be bigger than 1", ctx.Flags.Days)
		}
		remotes := ctx.Flags.Remote
		var err error
		if len(remotes) == 0 {
			remotes, err = core.ListRemoteNames()
			if err != nil {
				return errors.Trace(err, "list remote")
			}
		}

		store, err := core.NewRepositoryStorage()
		if err != nil {
			return errors.Trace(err, "init repo storage")
		}

		var items []*CleanItem
		repoPaths := make([]string, 0, len(remotes))
		for _, remoteName := range remotes {
			remote, err := core.GetRemote(remoteName)
			if err != nil {
				return errors.Trace(err, "get remote %q", remoteName)
			}
			repos := store.List(remoteName)
			for _, repo := range repos {
				repoPaths = append(repoPaths, repo.Path)
				var deltaDays int = -1
				if repo.LastAccess > 0 {
					if ctx.Flags.Never {
						continue
					}
					deltaSeconds := config.Now() - repo.LastAccess
					deltaDaysFlt := float64(deltaSeconds) / float64(config.DaySeconds)
					deltaDays = int(deltaDaysFlt)
					if deltaDays < ctx.Flags.Days {
						continue
					}
				}

				item := &CleanItem{
					Repo:   repo,
					Remote: remote,
					Days:   deltaDays,
				}
				if deltaDays < 0 {
					item.Never = true
				}
				items = append(items, item)
			}
		}

		ctx.OnClose(func() error { return store.Close() })
		ctx.Data = &CleanData{
			Items:     items,
			Store:     store,
			RepoPaths: repoPaths,
		}
		return nil
	},

	Run: func(ctx *app.Context[CleanFlags, CleanData]) error {
		items := ctx.Data.Items
		var err error
		if len(items) > 0 {
			if ctx.Flags.Edit {
				items, err = term.EditItems(config.Get().Editor,
					items, func(item *CleanItem) string {
						return item.Repo.FullName()
					})
				if err != nil {
					return err
				}
			}

			showCleanItems(items)
			if term.Confirm("continue") {
				for _, item := range items {
					err = ctx.Data.Store.DeleteAll(item.Repo)
					if err != nil {
						return err
					}
				}
				term.PrintOperation("clean repo done")
			}
		} else {
			term.PrintOperation("no repo to clean")
		}

		emptyDirs, err := osutil.ListEmptyDir(config.Get().Workspace, ctx.Data.RepoPaths)
		if err != nil {
			return err
		}
		if len(emptyDirs) > 0 {
			dirWord := english.Plural(len(emptyDirs), "dir", "dirs")
			term.Printf("%s to remove:", dirWord)
			for _, dir := range emptyDirs {
				term.Printf("* %s", dir)
			}
			if term.Confirm("continue") {
				for _, dir := range emptyDirs {
					err = os.RemoveAll(dir)
					if err != nil {
						return errors.Trace(err, "remove dir %s", dir)
					}
				}
				term.PrintOperation("clean empty dir done")
			}
		} else {
			term.PrintOperation("no empty dir to clean")
		}

		return nil
	},
})

func showCleanItems(items []*CleanItem) {
	nevers := make([]*CleanItem, 0)
	visited := make([]*CleanItem, 0, len(items))
	nameLen := 0
	for _, item := range items {
		if len(item.Repo.FullName()) > nameLen {
			nameLen = len(item.Repo.FullName())
		}
		if item.Never {
			nevers = append(nevers, item)
			continue
		}
		visited = append(visited, item)
	}

	sort.Slice(visited, func(i, j int) bool {
		return visited[i].Days > visited[j].Days
	})
	items = append(nevers, visited...)

	nameFmt := "%-" + strconv.Itoa(nameLen) + "s"
	itemWord := english.Plural(len(items), "repo", "repos")
	term.Printf("%s to delete:", itemWord)
	for _, item := range items {
		name := fmt.Sprintf(nameFmt, item.Repo.FullName())
		var view string
		if item.Never {
			view = term.Style("Never", "red")
		} else {
			daysWord := english.Plural(item.Days, "day", "days")
			view = term.Style(daysWord, "yellow")
		}
		term.Printf("* %s %s", name, view)
	}
}

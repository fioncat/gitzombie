package repo

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
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
		for _, remoteName := range remotes {
			remote, err := core.GetRemote(remoteName)
			if err != nil {
				return errors.Trace(err, "get remote %q", remoteName)
			}
			repos := store.List(remoteName)
			for _, repo := range repos {
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
			Items: items,
			Store: store,
		}
		return nil
	},

	Run: func(ctx *app.Context[CleanFlags, CleanData]) error {
		items := ctx.Data.Items
		if len(items) == 0 {
			term.Print("nothing to do")
			return nil
		}
		var err error
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
		term.ConfirmExit("continue")

		for _, item := range items {
			err := ctx.Data.Store.DeleteAll(item.Repo)
			if err != nil {
				return err
			}
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
	term.Print("%s to delete:", itemWord)
	for _, item := range items {
		name := fmt.Sprintf(nameFmt, item.Repo.FullName())
		var view string
		if item.Never {
			view = "red|Never|"
		} else {
			daysWord := english.Plural(item.Days, "day", "days")
			view = fmt.Sprintf("yellow|%s|", daysWord)
		}
		term.Print("* %s %s", name, view)
	}
}

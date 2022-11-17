package repo

import (
	"fmt"
	"strings"

	_ "embed"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type MergeFlags struct {
	Upstream bool

	TargetBranch string
	SourceBranch string
}

var Merge = app.Register(&app.Command[MergeFlags, Data]{
	Use:  "merge [-u] [-s source-branch] [-t target-branch]",
	Desc: "Create or open MergeRequest (PR in Github)",

	Init: initData[MergeFlags],

	Prepare: func(cmd *cobra.Command, flags *MergeFlags) {
		cmd.Flags().BoolVarP(&flags.Upstream, "upstream", "u", false, "merge to upstream repo")

		cmd.Flags().StringVarP(&flags.SourceBranch, "source", "s", "", "source branch")
		cmd.Flags().StringVarP(&flags.TargetBranch, "target", "t", "", "target branch")
		cmd.RegisterFlagCompletionFunc("source", app.Comp(app.CompGitLocalBranch(false)))
		cmd.RegisterFlagCompletionFunc("target", app.Comp(app.CompGitLocalBranch(false)))

		cmd.Args = cobra.ExactArgs(0)
	},

	Run: func(ctx *app.Context[MergeFlags, Data]) error {
		ctx.Data.Store.ReadOnly()
		err := git.EnsureNoUncommitted(git.Default)
		if err != nil {
			return err
		}

		repo, err := getCurrent(ctx)
		if err != nil {
			return err
		}

		apiRepo, err := apiGet(ctx, repo)
		if err != nil {
			return err
		}

		opts, err := mergeBuildOptions(ctx, apiRepo)
		if err != nil {
			return err
		}

		var url string
		err = execProvider("get merge", ctx.Data.Remote, func(p api.Provider) error {
			url, err = p.GetMerge(repo, *opts)
			return err
		})
		if err != nil {
			return err
		}
		if url != "" {
			return term.Open(url)
		}

		term.ConfirmExit("cannot find merge, do you want to create one")
		title, body, err := mergeEdit()
		if err != nil {
			return err
		}
		opts.Title = title
		opts.Body = body

		term.Print("")
		term.Print("About to create merge:")
		mergeShowInfo(repo, opts)
		term.ConfirmExit("continue")

		err = execProvider("create merge", ctx.Data.Remote, func(p api.Provider) error {
			url, err = p.CreateMerge(repo, *opts)
			return err
		})
		if err != nil {
			return err
		}

		return term.Open(url)
	},
})

func mergeBuildOptions(ctx *app.Context[MergeFlags, Data], apiRepo *api.Repository) (*api.MergeOption, error) {
	tar := ctx.Flags.TargetBranch
	if tar == "" {
		tar = apiRepo.DefaultBranch
	}

	src := ctx.Flags.SourceBranch
	if src == "" {
		cur, err := git.GetCurrentBranch(&git.Options{
			QuietCmd: true,
		})
		if err != nil {
			return nil, err
		}
		src = cur
	}

	var up *api.Repository
	if ctx.Flags.Upstream {
		if apiRepo.Upstream == nil {
			return nil, fmt.Errorf("repo %s does not have an upstream", apiRepo.Name)
		}
		up = apiRepo.Upstream
	}

	return &api.MergeOption{
		SourceBranch: src,
		TargetBranch: tar,
		Upstream:     up,
	}, nil
}

//go:embed merge_edit.md
var mergeEditContent string

const (
	mergeEditCommentStart = "<!--"
	mergeEditCommentEnd   = "-->"
)

func mergeEdit() (string, string, error) {
	content, err := term.EditContent(config.Get().Editor, mergeEditContent, "merge.md")
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(content, "\n")
	var (
		scanTitle   bool = true
		scanDiscard bool

		titleLines []string
		bodyLines  []string
	)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case scanTitle:
			if strings.HasPrefix(line, mergeEditCommentStart) {
				scanTitle = false
				scanDiscard = true
				continue
			}
			titleLines = append(titleLines, line)

		case scanDiscard:
			if strings.HasPrefix(line, mergeEditCommentEnd) {
				scanDiscard = false
				continue
			}

		default:
			bodyLines = append(bodyLines, line)
		}
	}
	title := strings.Join(titleLines, " ")
	title = strings.TrimSpace(title)
	if title == "" {
		return "", "", errors.New("merge title cannot be empty")
	}
	body := strings.Join(bodyLines, "\n")
	body = strings.TrimSpace(body)
	return title, body, nil
}

func mergeShowInfo(repo *core.Repository, opts *api.MergeOption) {
	var lineDesc string
	if opts.Body == "" {
		lineDesc = "yellow|empty|"
	} else {
		lineCount := len(strings.Split(opts.Body, "\n"))
		countWord := english.Plural(lineCount, "line", "")
		lineDesc = fmt.Sprintf("green|%s|", countWord)
	}
	var (
		src string
		tar string
	)
	if opts.Upstream == nil {
		src = fmt.Sprintf("green|%s|", opts.SourceBranch)
		tar = fmt.Sprintf("green|%s|", opts.TargetBranch)
	} else {
		src = fmt.Sprintf("magenta|%s|:greeen|%s|", repo.Name, opts.SourceBranch)
		tar = fmt.Sprintf("magenta|%s|:greeen|%s|", opts.Upstream.Name, opts.TargetBranch)
	}
	term.Print(" * Title:  green|%s|", opts.Title)
	term.Print(" * Body:   %s", lineDesc)
	term.Print(" * Source: %s", src)
	term.Print(" * Target: %s", tar)
}

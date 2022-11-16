package repo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/api"
	"github.com/fioncat/gitzombie/cmd/common"
	"github.com/fioncat/gitzombie/cmd/gitops"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type Merge struct {
	upstream bool

	sourceBranch string
}

func (m *Merge) Use() string    { return "merge [-u] [-s source-branch] [target-branch]" }
func (m *Merge) Desc() string   { return "Create or open MergeRequest (PR in Github)" }
func (m *Merge) Action() string { return "" }

func (m *Merge) Prepare(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&m.upstream, "upstream", "u", false, "merge to upstream repo")

	cmd.Flags().StringVarP(&m.sourceBranch, "source", "s", "", "source branch")
	cmd.RegisterFlagCompletionFunc("source", common.Comp(gitops.CompLocalBranch))

	cmd.Args = cobra.MaximumNArgs(1)
}

func (m *Merge) Run(ctx *Context, args common.Args) error {
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

	opts, err := m.buildOptions(apiRepo, args)
	if err != nil {
		return err
	}

	var url string
	err = execProvider("get merge", ctx.remote, func(p api.Provider) error {
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
	title, body, err := m.editTitleAndBody(opts)
	if err != nil {
		return err
	}
	opts.Title = title
	opts.Body = body

	term.Print("")
	term.Print("About to create merge:")
	m.showInfo(repo, opts)
	term.ConfirmExit("continue")

	err = execProvider("create merge", ctx.remote, func(p api.Provider) error {
		url, err = p.CreateMerge(repo, *opts)
		return err
	})
	if err != nil {
		return err
	}

	return term.Open(url)
}

func (m *Merge) buildOptions(apiRepo *api.Repository, args common.Args) (*api.MergeOption, error) {
	tar := args.Get(0)
	if tar == "" {
		tar = apiRepo.DefaultBranch
	}

	src := m.sourceBranch
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
	if m.upstream {
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

const mergeEditContent = `<!-- Please edit title and body for the new Merge.
Merge Info:
* Source Branch: %s
* Target Branch: %s
* Is Upstream:   %v

This file use markdown syntax, we will treat h1 title (starts with '#') as Merge's title.
The rest content above h1 title (not include title itself) will be treated as Merge's body.

After editing done, please quit this editor to continue.
-->

#
`

func (m *Merge) editTitleAndBody(opts *api.MergeOption) (string, string, error) {
	content := fmt.Sprintf(mergeEditContent, opts.SourceBranch,
		opts.TargetBranch, opts.Upstream != nil)
	content, err := term.EditContent(config.Get().Editor, content, "merge.md")
	if err != nil {
		return "", "", err
	}

	lines := strings.Split(content, "\n")
	var (
		title     string
		bodyLines []string

		scanBody bool
	)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") && !scanBody {
			title = strings.TrimPrefix(line, "#")
			scanBody = true
			continue
		}
		if scanBody {
			bodyLines = append(bodyLines, line)
		}
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return "", "", errors.New("merge title cannot be empty")
	}
	body := strings.Join(bodyLines, "\n")
	body = strings.TrimSpace(body)

	return title, body, nil
}

func (m *Merge) showInfo(repo *core.Repository, opts *api.MergeOption) {
	lineCount := len(strings.Split(opts.Body, "\n"))
	countWord := english.Plural(lineCount, "line", "")
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
	term.Print(" * Body:   green|%s|", countWord)
	term.Print(" * Source: %s", src)
	term.Print(" * Target: %s", tar)
}

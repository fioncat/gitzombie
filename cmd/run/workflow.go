package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/core"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/fioncat/gitzombie/pkg/worker"
	"github.com/spf13/cobra"
)

type WorkflowFlags struct {
	Edit bool

	Current bool

	LogPath string
}

var Workflow = app.Register(&app.Command[WorkflowFlags, app.Empty]{
	Use:    "workflow [-y] [-c] {workflow}",
	Desc:   "Run workflow",
	Action: "Run",

	Prepare: func(cmd *cobra.Command, flags *WorkflowFlags) {
		cmd.Flags().BoolVarP(&flags.Current, "current", "c", false, "run workflow on current repo")
		cmd.Flags().BoolVarP(&flags.Edit, "edit", "e", false, "edit repo to run")
		cmd.Flags().StringVarP(&flags.LogPath, "log-path", "", "", "log file path")

		cmd.Args = cobra.ExactArgs(1)
		cmd.ValidArgsFunction = app.Comp(app.CompWorkflow)
	},

	Run: func(ctx *app.Context[WorkflowFlags, app.Empty]) error {
		wf, err := core.GetWorkflow(ctx.Arg(0))
		if err != nil {
			return err
		}
		store, err := core.NewRepositoryStorage()
		if err != nil {
			return err
		}
		store.ReadOnly()

		if ctx.Flags.Current {
			return workflowRunCurrent(store, wf)
		}

		if wf.Select == nil {
			term.Print("nothing to do")
			return nil
		}

		items, err := wf.Select.Match(store)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			term.Print("no repo selected")
			return nil
		}
		if ctx.Flags.Edit {
			items, err = workflowEditItems(items)
			if err != nil {
				return err
			}
		}

		repoWord := english.Plural(len(items), "repo", "repos")
		term.ConfirmExit("Do you want to run workflow blue|%s| on %s", ctx.Arg(0), repoWord)

		return workflowRun(ctx, wf.Jobs, items)
	},
})

func workflowRunCurrent(store *core.RepositoryStorage, wf *core.Workflow) error {
	path, err := os.Getwd()
	if err != nil {
		return errors.Trace(err, "get pwd")
	}
	env := make(osutil.Env)
	repo, _ := store.GetByPath(path)
	if repo != nil {
		var remote *core.Remote
		if repo.Remote != "" {
			remote, err = core.GetRemote(repo.Remote)
			if err != nil {
				return err
			}
		}
		err = repo.SetEnv(remote, env)
		if err != nil {
			return errors.Trace(err, "set env for repo %s", repo.FullName())
		}
	}
	for _, job := range wf.Jobs {
		err = job.Execute(path, env)
		if err != nil {
			return err
		}
	}
	return nil
}

func workflowEditItems(items []*core.WorkflowMatchItem) ([]*core.WorkflowMatchItem, error) {
	itemMap := make(map[string]*core.WorkflowMatchItem, len(items))
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = item.Path
		itemMap[item.Path] = item
	}
	content := strings.Join(lines, "\n")
	content, err := term.EditContent(config.Get().Editor, content, "items.txt")
	if err != nil {
		return nil, err
	}
	lines = strings.Split(content, "\n")

	editedItems := make([]*core.WorkflowMatchItem, 0, len(lines))
	set := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name := line
		item := itemMap[name]
		if item == nil {
			return nil, fmt.Errorf("edit: cannot find item %q", name)
		}
		if _, ok := set[name]; ok {
			continue
		}
		editedItems = append(editedItems, item)
		set[name] = struct{}{}
	}
	if len(editedItems) == 0 {
		return nil, errors.New("nothing to do after editing")
	}
	return editedItems, nil
}

func workflowRun(ctx *app.Context[WorkflowFlags, app.Empty], jobs []*core.Job, items []*core.WorkflowMatchItem) error {
	tasks := make([]*worker.Task[core.WorkflowMatchItem], len(items))
	for i, item := range items {
		tasks[i] = &worker.Task[core.WorkflowMatchItem]{
			Name:  item.Path,
			Value: item,
		}
	}

	w := worker.New("running", tasks)

	core.MuteJob = true
	errs := w.Run(func(_ string, item *core.WorkflowMatchItem) error {
		for _, job := range jobs {
			err := job.Execute(item.Path, item.Env)
			if err != nil {
				return err
			}
		}
		return nil
	})
	term.Print("")

	if len(errs) > 0 {
		var sb strings.Builder
		for i, err := range errs {
			if jobErr, ok := err.(*core.JobError); ok {
				header := fmt.Sprintf("=======> output of %q on %q: %v\n",
					jobErr.Name, jobErr.Path, jobErr.Err)
				sb.WriteString(header)
				sb.WriteString(jobErr.Out)
			} else {
				content := fmt.Sprintf("=======> #%d error: %v\n", i, err)
				sb.WriteString(content)
			}
			sb.WriteString("\n\n")
		}
		logPath := ctx.Flags.LogPath
		if logPath == "" {
			dateStr := time.Now().Format("2006-01-02_15:04:05")
			logName := fmt.Sprintf("%s-%s", ctx.Arg(0), dateStr)
			logPath = filepath.Join(os.TempDir(), "gitzombie", "workflows", logName)
		}
		err := osutil.WriteFile(logPath, []byte(sb.String()))
		if err != nil {
			return errors.Trace(err, "write log file")
		}
		errWord := english.Plural(len(errs), "error", "")
		term.Print("write red|%s log| to green|%s|", errWord, logPath)
		return fmt.Errorf("workflow failed with %s", errWord)
	}

	term.Print("workflow blue|%s| done", ctx.Arg(0))
	return nil
}

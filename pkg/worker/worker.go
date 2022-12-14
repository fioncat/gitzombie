package worker

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

var Count = runtime.NumCPU()

type Action[T any] func(task *Task[T]) error

type Task[T any] struct {
	Name  string
	Value *T

	Action Action[T]

	done bool
	fail bool
}

type Tracker[T any] interface {
	Render(total int) func()

	Add(task *Task[T])
}

type taskError[T any] struct {
	err  error
	task *Task[T]
}

type Worker[T any] struct {
	Name string

	Tasks   []*Task[T]
	Tracker Tracker[T]

	LogPath string
}

func (w *Worker[T]) Run(action Action[T]) error {
	if Count <= 0 {
		Count = runtime.NumCPU()
	}
	taskChan := make(chan *Task[T], len(w.Tasks))
	errChan := make(chan *taskError[T], len(w.Tasks))

	waitRender := w.Tracker.Render(len(w.Tasks))

	var wg sync.WaitGroup
	wg.Add(Count)
	for i := 0; i < Count; i++ {
		go func() {
			defer wg.Done()
			for task := range taskChan {
				h := action
				if task.Action != nil {
					h = task.Action
				}
				w.Tracker.Add(task)
				err := h(task)
				if err != nil {
					task.fail = true
					errChan <- &taskError[T]{
						task: task,
						err:  err,
					}
				}
				task.done = true
			}
		}()
	}
	for _, task := range w.Tasks {
		taskChan <- task
	}
	close(taskChan)
	wg.Wait()
	close(errChan)
	var errs []*taskError[T]
	for err := range errChan {
		errs = append(errs, err)
	}
	waitRender()

	if len(errs) > 0 {
		return w.handleErrors(errs)
	}
	return nil
}

type outError interface {
	Out() string
}

func (w *Worker[T]) handleErrors(errs []*taskError[T]) error {
	var sb bytes.Buffer
	sb.Grow(len(errs))
	for _, err := range errs {
		header := fmt.Sprintf("=> handle %s failed: %v\n", err.task.Name, err.err)
		sb.WriteString(header)
		if outErr, ok := err.err.(outError); ok {
			sb.WriteString(outErr.Out())
		}
		sb.WriteString("\n")
	}
	logPath := w.LogPath
	if logPath == "" {
		logPath = filepath.Join(os.TempDir(), "gitzombie", "logs", w.Name)
	}
	err := osutil.WriteFile(logPath, sb.Bytes())
	if err != nil {
		return errors.Trace(err, "write log file")
	}
	errWord := english.Plural(len(errs), "error", "")
	term.Println()
	term.Printf("write %s log to %s", term.Style(errWord, "red"), logPath)
	return fmt.Errorf("%s failed with %s", w.Name, errWord)
}

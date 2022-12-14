package worker

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	renderInterval = time.Millisecond * 250
)

type jobTracker[T any] struct {
	lock sync.Mutex

	running []*Task[T]

	lastRunningCount int

	total int

	doneFmt string
	done    int

	verb string
}

func NewJobTracker[T any](verb string) Tracker[T] {
	return &jobTracker[T]{
		verb: verb,
	}
}

func (t *jobTracker[T]) Render(total int) func() {
	totalStr := strconv.Itoa(total)
	totalLen := len(totalStr)
	doneFmt := "%" + strconv.Itoa(totalLen) + "d"

	t.doneFmt = doneFmt
	t.total = total

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range time.Tick(renderInterval) {
			t.render()
			if t.done >= t.total {
				return
			}
		}
	}()
	return func() {
		<-done
	}
}

func (t *jobTracker[T]) render() {
	t.lock.Lock()
	defer t.lock.Unlock()

	dones := make([]*Task[T], 0)
	running := make([]*Task[T], 0, len(t.running))
	for _, task := range t.running {
		if task.done {
			dones = append(dones, task)
			continue
		}
		running = append(running, task)
	}
	t.running = running

	var out strings.Builder
	out.Grow(t.lastRunningCount)

	for t.lastRunningCount > 0 {
		cursorUp(&out)
		t.lastRunningCount--
	}
	t.lastRunningCount = len(running)

	for _, task := range dones {
		t.done++
		doneStr := fmt.Sprintf(t.doneFmt, t.done)
		var line string
		if task.fail {
			line = fmt.Sprintf("(%s/%d) %s failed\n", doneStr, t.total, task.Name)
			line = term.Style(line, "red")
		} else {
			head := fmt.Sprintf("(%s/%d)", doneStr, t.total)
			head = term.Style(head, "bold")
			line = fmt.Sprintf("%s %s done\n", head, task.Name)
		}
		out.WriteString(line)
	}

	verbHead := term.Style(t.verb, "yellow")
	for _, task := range running {
		line := fmt.Sprintf("%s %s\n", verbHead, task.Name)
		out.WriteString(line)
	}
	fmt.Fprint(os.Stderr, out.String())
}

func (t *jobTracker[T]) Add(task *Task[T]) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.running = append(t.running, task)
}

func cursorUp(out *strings.Builder) {
	out.WriteString(text.CursorUp.Sprint())
	out.WriteString(text.EraseLine.Sprint())
}

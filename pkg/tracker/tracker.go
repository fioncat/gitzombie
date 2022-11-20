package tracker

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

type Tracker struct {
	lock sync.Mutex

	running []*task

	lastRunningCount int

	doneFmt string
	done    int

	total int

	verb string
}

type task struct {
	name string

	done bool
	fail bool
}

func New(verb string, total int) *Tracker {
	totalStr := strconv.Itoa(total)
	totalLen := len(totalStr)
	doneFmt := "%" + strconv.Itoa(totalLen) + "d"
	return &Tracker{
		total:   total,
		doneFmt: doneFmt,
		verb:    verb,
	}
}

func (t *Tracker) Render() func() {
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

func (t *Tracker) render() {
	t.lock.Lock()
	defer t.lock.Unlock()

	dones := make([]*task, 0)
	running := make([]*task, 0, len(t.running))
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
		out.WriteString(text.CursorUp.Sprint())
		out.WriteString(text.EraseLine.Sprint())
		t.lastRunningCount--
	}
	t.lastRunningCount = len(running)

	for _, task := range dones {
		t.done++
		doneStr := fmt.Sprintf(t.doneFmt, t.done)
		var line string
		if task.fail {
			line = fmt.Sprintf("red|(%s/%d)| %s failed\n", doneStr, t.total, task.name)
		} else {
			line = fmt.Sprintf("green|(%s/%d)| %s done\n", doneStr, t.total, task.name)
		}
		out.WriteString(line)
	}

	for _, task := range running {
		line := fmt.Sprintf("yellow|%s| %s\n", t.verb, task.name)
		out.WriteString(line)
	}
	lines := term.Color(out.String())
	fmt.Fprint(os.Stderr, lines)
}

func (t *Tracker) Start(name string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.running = append(t.running, &task{name: name})
}

func (t *Tracker) Done(name string, ok bool) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, task := range t.running {
		if task.name == name {
			task.done = true
			task.fail = !ok
			return
		}
	}
}

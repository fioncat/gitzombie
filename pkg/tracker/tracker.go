package tracker

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	renderInterval = time.Millisecond * 50
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
	if t.lastRunningCount > 0 {
		fmt.Fprint(os.Stderr, text.CursorUp.Sprintn(t.lastRunningCount))
	}
	t.lastRunningCount = len(running)

	for _, task := range dones {
		t.done++
		doneStr := fmt.Sprintf(t.doneFmt, t.done)
		fmt.Fprint(os.Stderr, text.EraseLine.Sprintn(t.lastRunningCount))
		if task.fail {
			term.Print("red|(%s/%d)| %s failed", doneStr, t.total, task.name)
		} else {
			term.Print("green|(%s/%d)| %s done", doneStr, t.total, task.name)
		}
	}

	for _, task := range running {
		fmt.Fprint(os.Stderr, text.EraseLine.Sprintn(t.lastRunningCount))
		term.Print("yellow|%s| %s", t.verb, task.name)
	}
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

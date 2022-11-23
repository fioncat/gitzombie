package worker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

type BytesTask struct {
	// The task start time, using to calculate download/upload speed.
	start int64

	reader io.ReadCloser  // read source
	writer io.WriteCloser // write source

	size  uint64 // totally bytes size
	speed uint64 // bytes hanlde speed

	total uint64

	wait bool
}

func DownloadTask(name string, reader io.ReadCloser, total uint64) *Task[BytesTask] {
	return &Task[BytesTask]{
		Name: name,
		Value: &BytesTask{
			total:  total,
			wait:   true,
			reader: reader,
		},
	}
}

func (bt *BytesTask) Write(data []byte) (int, error) {
	size, err := bt.writer.Write(data)
	if err != nil {
		return 0, err
	}
	bt.grow(len(data))
	return size, nil
}

func (bt *BytesTask) Read(p []byte) (int, error) {
	size, err := bt.reader.Read(p)
	if err != nil {
		return 0, err
	}
	bt.grow(len(p))
	return size, nil
}

func (bt *BytesTask) grow(delta int) {
	now := time.Now().Unix()
	newSize := atomic.AddUint64(&bt.size, uint64(delta))
	took := now - bt.start
	if took > 0 {
		speedFloat := float64(newSize) / float64(took)
		atomic.StoreUint64(&bt.speed, uint64(speedFloat))
	}
}

type BytesTracker struct {
	lock sync.Mutex

	tasks []*Task[BytesTask]

	nameFmt string

	rendered bool

	verb    string
	verbing string
}

func NewBytesTracker(tasks []*Task[BytesTask], verb, verbing string) Tracker[BytesTask] {
	return &BytesTracker{
		tasks:   tasks,
		verb:    verb,
		verbing: verbing,
	}
}

func (t *BytesTracker) Render(_ int) func() {
	var nameLen int
	for _, task := range t.tasks {
		if len(task.Name) > nameLen {
			nameLen = len(task.Name)
		}
	}
	nameFmt := "%-" + strconv.Itoa(nameLen) + "s"
	t.nameFmt = nameFmt

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range time.Tick(renderInterval) {
			if t.render() {
				return
			}
		}
	}()
	return func() {
		<-done
	}
}

func (t *BytesTracker) render() bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	var out strings.Builder
	out.Grow(len(t.tasks))

	if t.rendered {
		for i := 0; i < len(t.tasks); i++ {
			cursorUp(&out)
		}
	} else {
		t.rendered = true
	}

	done := true
	for _, task := range t.tasks {
		name := fmt.Sprintf(t.nameFmt, task.Name)
		var line string
		if task.done {
			if task.fail {
				line = fmt.Sprintf("%s %s red|failed|\n", t.verb, task.Name)
			} else {
				line = fmt.Sprintf("%s %s done\n", t.verb, task.Name)
			}
		} else {
			if task.Value.wait {
				line = fmt.Sprintf("%s %s waitting...\n", t.verb, name)
			} else {
				size := bytesStr(atomic.LoadUint64(&task.Value.size))
				speed := bytesStr(atomic.LoadUint64(&task.Value.speed))
				line = fmt.Sprintf("%s %s... %s (%s/s)\n", t.verbing, task.Name, size, speed)
			}
			done = false
		}
		out.WriteString(line)
	}

	lines := term.Color(out.String())
	fmt.Fprint(os.Stderr, lines)
	return done
}

func (t *BytesTracker) Add(task *Task[BytesTask]) {}

func bytesStr(b uint64) string {
	size := humanize.IBytes(b)
	return strings.ReplaceAll(size, " ", "")
}

type Bytes struct {
	Tracker Tracker[BytesTask]

	Tasks []*Task[BytesTask]

	LogPath string
}

func (b *Bytes) Download(dir string) error {
	err := osutil.EnsureDir(dir)
	if err != nil {
		return errors.Trace(err, "ensure dir")
	}

	now := time.Now().Unix()
	for _, task := range b.Tasks {
		if task.Value.reader == nil {
			// reader is the download source, it cannot be empty.
			panic("when using download, please set reader for task")
		}
		path := filepath.Join(dir, task.Name)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Trace(err, "open file %q", task.Name)
		}
		task.Action = func(task *Task[BytesTask]) error {
			defer file.Close()
			defer task.Value.reader.Close()

			task.Value.wait = false
			_, err := io.Copy(file, task.Value)
			return err
		}

		task.Value.start = now
	}

	w := &Worker[BytesTask]{
		Name:    "download",
		Tasks:   b.Tasks,
		Tracker: b.Tracker,
		LogPath: b.LogPath,
	}
	return w.Run(nil)
}

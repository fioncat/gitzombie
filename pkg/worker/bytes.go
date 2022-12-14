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
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type BytesTask struct {
	// The task start time, using to calculate download/upload speed.
	start int64

	reader io.ReadCloser  // read source
	writer io.WriteCloser // write source

	size  uint64 // totally bytes size
	speed uint64 // bytes hanlde speed

	total uint64

	totalStr string

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

func setBytesTaskTotalStr(tasks []*Task[BytesTask]) {
	sizeLen := 0
	for _, task := range tasks {
		size := bytesStr(task.Value.total)
		if len(size) > sizeLen {
			sizeLen = len(size)
		}
		task.Value.totalStr = size
	}
	totalFmt := "%" + strconv.Itoa(sizeLen) + "s"
	for _, task := range tasks {
		task.Value.totalStr = fmt.Sprintf(totalFmt, task.Value.totalStr)
	}
}

type BytesTracker struct {
	lock sync.Mutex

	tasks []*Task[BytesTask]

	nameFmt string

	rendered bool
}

func NewBytesTracker(tasks []*Task[BytesTask]) Tracker[BytesTask] {
	return &BytesTracker{
		tasks: tasks,
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
				line = fmt.Sprintf("%s %s\n", task.Name, term.Style("failed", "red"))
			} else {
				line = fmt.Sprintf("%s done\n", task.Name)
			}
		} else {
			if task.Value.wait {
				line = fmt.Sprintf("%s waitting...\n", name)
			} else {
				size := bytesStr(atomic.LoadUint64(&task.Value.size))
				speed := bytesStr(atomic.LoadUint64(&task.Value.speed))
				line = fmt.Sprintf("%s... %s (%s/s)\n", task.Name, size, speed)
			}
			done = false
		}
		out.WriteString(line)
	}

	fmt.Fprint(os.Stderr, out.String())
	return done
}

func (t *BytesTracker) Add(task *Task[BytesTask]) {}

func bytesStr(b uint64) string {
	size := humanize.IBytes(b)
	return strings.ReplaceAll(size, " ", "")
}

type BytesBarTracker struct {
	proc *mpb.Progress

	tasks []*Task[BytesTask]
	bars  []*mpb.Bar
}

func NewBytesBarTracker(tasks []*Task[BytesTask]) Tracker[BytesTask] {
	t := &BytesBarTracker{proc: mpb.New()}
	for _, task := range tasks {
		t.initTask(task)
	}
	return t
}

func (t *BytesBarTracker) Render(_ int) func() {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range time.Tick(renderInterval) {
			if t.loop() {
				return
			}
		}
	}()
	return func() {
		<-done
		t.proc.Wait()
	}
}

func (t *BytesBarTracker) loop() bool {
	done := true
	for i, task := range t.tasks {
		bar := t.bars[i]
		cur := task.Value.size
		if task.done {
			cur = task.Value.total
		} else {
			done = false
		}
		bar.SetCurrent(int64(cur))
	}
	return done
}

func (t *BytesBarTracker) Add(task *Task[BytesTask]) {}

func (t *BytesBarTracker) initTask(task *Task[BytesTask]) {
	bar := t.proc.AddBar(int64(task.Value.total),
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name(task.Name+strings.Repeat(" ", 10), decor.WCSyncWidthR),
			decor.OnComplete(decor.CurrentKibiByte("%2d"), ""),
			decor.OnComplete(decor.AverageSpeed(decor.UnitKiB, " [%d] "), ""),
			decor.OnComplete(decor.AverageETA(decor.ET_STYLE_MMSS), ""),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(), ""),
			decor.Any(func(s decor.Statistics) string {
				if s.Completed {
					if task.fail {
						return term.Style("failed", "red")
					}
					return fmt.Sprintf("%s done", task.Value.totalStr)
				}
				return ""
			}),
		),
	)
	t.bars = append(t.bars, bar)
	t.tasks = append(t.tasks, task)
}

type Bytes struct {
	Tracker Tracker[BytesTask]

	Tasks []*Task[BytesTask]

	LogPath string
}

func (b *Bytes) Download(dir string) error {
	if dir != "" {
		err := osutil.EnsureDir(dir)
		if err != nil {
			return errors.Trace(err, "ensure dir")
		}
	}

	setBytesTaskTotalStr(b.Tasks)

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

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

type downloadTask struct {
	start int64

	file *os.File
	size uint64

	speed uint64

	reader io.ReadCloser

	mu sync.Mutex

	wait bool
}

func (t *downloadTask) Write(data []byte) (int, error) {
	size, err := t.file.Write(data)
	if err != nil {
		return 0, err
	}
	newSize := atomic.AddUint64(&t.size, uint64(size))
	now := time.Now().Unix()
	took := now - t.start
	if took > 0 {
		speedFloat := float64(newSize) / float64(took)
		atomic.StoreUint64(&t.speed, uint64(speedFloat))
	}

	return size, nil
}

type downloadTracker struct {
	lock sync.Mutex

	tasks []*Task[downloadTask]

	nameFmt string

	rednered bool
}

func (t *downloadTracker) Render(_ int) func() {
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

func (t *downloadTracker) render() bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	var out strings.Builder
	out.Grow(len(t.tasks))

	if t.rednered {
		for i := 0; i < len(t.tasks); i++ {
			cursorUp(&out)
		}
	} else {
		t.rednered = true
	}

	done := true
	for _, task := range t.tasks {
		name := fmt.Sprintf(t.nameFmt, task.Name)
		var msg string
		if task.done {
			if task.fail {
				msg = fmt.Sprintf("Download %s red|failed|\n", name)
			} else {
				msg = fmt.Sprintf("Download %s done\n", name)
			}
		} else {
			if task.Value.wait {
				msg = fmt.Sprintf("Download %s waitting...\n", name)
			} else {
				size := bytesStr(atomic.LoadUint64(&task.Value.size))
				size = strings.ReplaceAll(size, " ", "")
				msg = fmt.Sprintf("Downloading %s... %s", name, size)
				if task.Value.speed > 0 {
					speed := atomic.LoadUint64(&task.Value.speed)
					speedSize := bytesStr(speed)
					msg = fmt.Sprintf("%s (%s/s)", msg, speedSize)
				}
				msg += "\n"
			}
			done = false
		}
		out.WriteString(msg)
	}

	lines := term.Color(out.String())
	fmt.Fprint(os.Stderr, lines)
	return done
}

func (t *downloadTracker) Add(task *Task[downloadTask]) {}

type DownloadTask struct {
	Name string

	Reader io.ReadCloser
}

type Downloader struct {
	worker *Worker[downloadTask]
}

func NewDownloader(dir string, tasks []*DownloadTask) (*Downloader, error) {
	err := osutil.EnsureDir(dir)
	if err != nil {
		return nil, errors.Trace(err, "ensure dir")
	}

	now := time.Now().Unix()
	convertTasks := make([]*Task[downloadTask], len(tasks))
	for i, task := range tasks {
		path := filepath.Join(dir, task.Name)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return nil, errors.Trace(err, "open download file")
		}
		convertTasks[i] = &Task[downloadTask]{
			Name: task.Name,
			Value: &downloadTask{
				start:  now,
				file:   file,
				reader: task.Reader,
				wait:   true,
			},
		}

	}

	w := &Worker[downloadTask]{
		Name: "download",

		Tasks: convertTasks,
		Tracker: &downloadTracker{
			tasks: convertTasks,
		},
	}
	return &Downloader{worker: w}, nil
}

func (w *Downloader) Run(logPath string) error {
	w.worker.LogPath = logPath
	return w.worker.Run(func(task *Task[downloadTask]) error {
		defer task.Value.file.Close()
		defer task.Value.reader.Close()
		task.Value.wait = false

		_, err := io.Copy(task.Value, task.Value.reader)
		return err
	})
}

func bytesStr(b uint64) string {
	size := humanize.IBytes(b)
	return strings.ReplaceAll(size, " ", "")
}

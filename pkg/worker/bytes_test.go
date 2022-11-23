package worker

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

type testReadCloser struct {
	reader io.Reader
}

func (r *testReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *testReadCloser) Close() error {
	return nil
}

func TestDownload(t *testing.T) {
	Count = 3
	tasks := make([]*Task[BytesTask], 10)
	for i := 0; i < 10; i++ {
		var total int64 = 1024 * 1024 * (50 + int64(i)*20)

		reader := io.LimitReader(rand.Reader, total)
		readCloser := &testReadCloser{reader: reader}

		name := fmt.Sprintf("test_%d", i)
		tasks[i] = DownloadTask(name, readCloser, uint64(total))
	}
	w := &Bytes{
		Tracker: NewBytesTracker(tasks, "Download", "Downloading"),
		Tasks:   tasks,
	}

	err := w.Download("tmp")
	if err != nil {
		t.Fatal(err)
	}
}

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
	tasks := make([]*DownloadTask, 3)
	for i := 0; i < 3; i++ {
		var total int64 = 1024 * 1024 * (500 + int64(i)*100)

		reader := io.LimitReader(rand.Reader, total)
		readCloser := &testReadCloser{reader: reader}

		tasks[i] = &DownloadTask{
			Name:   fmt.Sprintf("test_%d", i),
			Reader: readCloser,
		}
	}
	d, err := NewDownloader("tmp", tasks)
	if err != nil {
		t.Fatal(err)
	}

	err = d.Run("")
	if err != nil {
		t.Fatal(err)
	}
}

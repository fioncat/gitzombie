package worker

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

func TestProgress(t *testing.T) {
	pw := progress.NewWriter()
	pw.SetNumTrackersExpected(10)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerLength(50)
	// pw.SetNumTrackersExpected(10)

	pw.Style().Options.Separator = ""
	pw.Style().Options.DoneString = ""
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Visibility.ETA = false
	pw.Style().Visibility.ETAOverall = false
	pw.Style().Visibility.Speed = false
	pw.Style().Visibility.Pinned = true
	pw.Style().Visibility.Tracker = false
	pw.Style().Visibility.Percentage = false
	pw.Style().Visibility.Time = false
	pw.Style().Visibility.Value = false
	pw.Style().Visibility.TrackerOverall = false
	go pw.Render()

	var wg sync.WaitGroup
	var done int
	var doneLock sync.Mutex
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		wg.Add(1)
		tk := &progress.Tracker{
			Message: fmt.Sprintf("cloning fioncat/%d", i),
			Units:   progress.UnitsCurrencyDollar,
		}
		i := i
		go func() {
			pw.AppendTracker(tk)
			time.Sleep(time.Second * 3)
			doneLock.Lock()
			defer doneLock.Unlock()

			done++
			msg := fmt.Sprintf("(%d/10) clone fioncat/%d done", done, i)
			tk.UpdateMessage(msg)
			tk.MarkAsDone()
		}()
	}
	wg.Wait()
}

func TestWorker(t *testing.T) {
	type testValue struct {
		idx int
	}

	var total = 30
	tasks := make([]*Task[testValue], total)
	for i := 0; i < total; i++ {
		tasks[i] = &Task[testValue]{
			Name:  fmt.Sprintf("task %d", i),
			Value: &testValue{idx: i},
		}
	}

	w := New("testing", tasks)
	errs := w.Run(func(_ string, val *testValue) error {
		idx := val.idx
		switch idx {
		case 2, 5, 7:
			time.Sleep(time.Second * 5)
			return nil

		case 10, 13, 16:
			time.Sleep(time.Second * 4)
			return errors.New("test error")
		}
		time.Sleep(time.Second * 3)
		return nil
	})
	if len(errs) != 3 {
		t.Fatalf("unexpect error: %v", errs)
	}

}

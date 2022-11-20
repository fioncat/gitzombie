package worker

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

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

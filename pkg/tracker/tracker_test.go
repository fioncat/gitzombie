package tracker

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTracker(t *testing.T) {
	var total int = 20
	tracker := New("testing", total)
	renderDone := tracker.Render()

	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		time.Sleep(time.Second)
		i := i
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("task %d", i)
			if i%5 == 0 {
				name = name + " long"
				tracker.Start(name)
				time.Sleep(time.Second * 5)
			} else {
				tracker.Start(name)
				time.Sleep(time.Second * 2)
			}
			tracker.Done(name, true)
		}()
	}

	wg.Wait()
	renderDone()
}

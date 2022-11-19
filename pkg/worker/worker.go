package worker

import (
	"runtime"
	"sync"

	"github.com/fioncat/gitzombie/pkg/tracker"
)

var (
	Count = runtime.NumCPU()
)

type Action[T any] func(name string, val *T) error

type Task[T any] struct {
	Name   string
	Action Action[T]
	Value  *T
}

type Worker[T any] struct {
	tasks []*Task[T]

	verb string
}

func New[T any](verb string, tasks []*Task[T]) *Worker[T] {
	return &Worker[T]{
		tasks: tasks,
		verb:  verb,
	}
}

func (w *Worker[T]) Run(action Action[T]) []error {
	taskChan := make(chan *Task[T], len(w.tasks))
	errChan := make(chan error, len(w.tasks))

	tck := tracker.New(w.verb, len(w.tasks))
	waitRender := tck.Render()

	var wg sync.WaitGroup
	wg.Add(Count)
	for i := 0; i < Count; i++ {
		go func() {
			defer wg.Done()
			for task := range taskChan {
				h := action
				if task.Action != nil {
					h = task.Action
				}
				tck.Start(task.Name)
				err := h(task.Name, task.Value)
				tck.Done(task.Name, err == nil)
				if err != nil {
					errChan <- err
				}
			}
		}()
	}
	for _, task := range w.tasks {
		taskChan <- task
	}
	close(taskChan)
	wg.Wait()
	close(errChan)
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	waitRender()
	return errs
}

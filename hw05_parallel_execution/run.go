package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	workCh := make(chan Task)

	wg := sync.WaitGroup{}
	wg.Add(n)
	var z int32
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for task := range workCh {
				err := task()
				if err != nil {
					atomic.AddInt32(&z, 1)
				}
			}
		}()
	}
	for _, task := range tasks {
		if atomic.LoadInt32(&z) >= int32(m) {
			break
		}
		workCh <- task
	}

	close(workCh)
	wg.Wait()
	if z >= int32(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}

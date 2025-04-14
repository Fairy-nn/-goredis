package wait

import (
	"sync"
	"time"
)

type Wait struct {
	wg sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}
func (w *Wait) Done() {
	w.wg.Done()
}
func (w *Wait) Wait() {
	w.wg.Wait()
}
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	wc := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(wc)
	}()
	select {
	case <-wc:
		return true
	case <-time.After(timeout):
		return false
	}
}

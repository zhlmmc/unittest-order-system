package concurrent

import (
	"errors"
	"sync"
)

// Task represents a function that can be executed by the pool
type Task func() error

// Pool represents a worker pool
type Pool struct {
	workers    chan struct{}
	wg         sync.WaitGroup
	activeJobs int32
	closed     bool
	mu         sync.Mutex
}

// NewPool creates a new worker pool with the specified number of workers
func NewPool(size int) *Pool {
	return &Pool{
		workers: make(chan struct{}, size),
	}
}

// Submit submits a task to the pool
func (p *Pool) Submit(task func() error) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return errors.New("pool is closed")
	}
	p.mu.Unlock()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.workers <- struct{}{}        // acquire worker
		defer func() { <-p.workers }() // release worker
		_ = task()
	}()

	return nil
}

// Close closes the pool and waits for all tasks to complete
func (p *Pool) Close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	p.wg.Wait()
}

// ActiveTasks returns the number of active tasks
func (p *Pool) ActiveTasks() int {
	return len(p.workers)
}

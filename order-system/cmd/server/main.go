package main

import (
	"fmt"
	"time"

	"order-system/pkg/infra/concurrent"
	"order-system/pkg/infra/errors"
)

func main() {
	// Test error package
	err := errors.New("TEST_ERROR", "this is a test error")
	fmt.Printf("Error: %v\n", err)

	// Test counter
	counter := concurrent.NewCounter(0)
	counter.Increment()
	fmt.Printf("Counter value: %d\n", counter.Value())

	// Test pool
	pool := concurrent.NewPool(2)
	defer pool.Close()

	// Submit some tasks
	for i := 0; i < 5; i++ {
		i := i // capture loop variable
		err := pool.Submit(func() error {
			fmt.Printf("Executing task %d\n", i)
			time.Sleep(time.Second)
			return nil
		})
		if err != nil {
			fmt.Printf("Failed to submit task: %v\n", err)
		}
	}

	// Wait for tasks to complete
	time.Sleep(3 * time.Second)
	fmt.Printf("Active tasks: %d\n", pool.ActiveTasks())
}

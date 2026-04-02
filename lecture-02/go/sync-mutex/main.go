// go run main.go
//
// sync.Mutex — простейший мьютекс.
// Защищает критическую секцию: только одна горутина внутри Lock/Unlock.

package main

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *SafeCounter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func main() {
	var counter SafeCounter
	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}

	wg.Wait()
	fmt.Printf("Счётчик: %d (ожидается 1000)\n", counter.Get())
}

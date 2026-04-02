// go run main.go
//
// «Глупый» спинлок на test-and-set (CAS).
// Нечестный (unfair) — нет гарантии порядка захвата.
// Все горутины крутятся в busy-wait, нагружая шину кэш-когерентности.

package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Spinlock — простейший TAS-спинлок.
type Spinlock struct {
	locked atomic.Int32
}

func (s *Spinlock) Lock() {
	for !s.locked.CompareAndSwap(0, 1) {
		// spin — активное ожидание
	}
}

func (s *Spinlock) Unlock() {
	s.locked.Store(0)
}

func main() {
	runtime.GOMAXPROCS(4)

	var lock Spinlock
	sharedCounter := 0

	var wg sync.WaitGroup

	for id := range 4 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runtime.LockOSThread()

			for range 100 {
				lock.Lock()
				fmt.Printf("Thread %d\n", id)
				sharedCounter++
				time.Sleep(10 * time.Millisecond)
				lock.Unlock()
			}
		}(id)
	}

	wg.Wait()
	fmt.Printf("Итоговое значение счётчика: %d (ожидается 400)\n", sharedCounter)
}

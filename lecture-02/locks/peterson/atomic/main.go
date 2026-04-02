// go run main.go
//
// Алгоритм Петерсона на sync/atomic.
// Работает корректно, потому что atomic операции дают необходимые memory barriers.

package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type PetersonMutex struct {
	flag [2]atomic.Int32
	turn atomic.Int32
}

func (m *PetersonMutex) Lock(self int) {
	other := 1 - self
	m.flag[self].Store(1)
	m.turn.Store(int32(other))
	for m.flag[other].Load() == 1 && m.turn.Load() == int32(other) {
		// spin
	}
}

func (m *PetersonMutex) Unlock(self int) {
	m.flag[self].Store(0)
}

func main() {
	runtime.GOMAXPROCS(4)

	var mu PetersonMutex
	sharedCounter := 0

	var wg sync.WaitGroup
	wg.Add(2)

	for id := range 2 {
		go func(id int) {
			defer wg.Done()
			runtime.LockOSThread()

			for range 100 {
				mu.Lock(id)
				fmt.Printf("Thread %d\n", id)
				sharedCounter++
				time.Sleep(10 * time.Millisecond)
				mu.Unlock(id)
			}
		}(id)
	}

	wg.Wait()
	fmt.Printf("Итоговое значение счётчика: %d (ожидается 200)\n", sharedCounter)
}

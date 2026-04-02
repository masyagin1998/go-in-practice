// go run main.go
//
// Алгоритм Петерсона БЕЗ atomic — сломанная версия.
// Без memory barriers CPU и компилятор переупорядочивают store/load,
// нарушая mutual exclusion. Счётчик будет < 200 или программа зависнет.

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type PetersonMutex struct {
	flag [2]int32
	turn int32
}

func (m *PetersonMutex) Lock(self int) {
	other := 1 - self
	m.flag[self] = 1
	m.turn = int32(other)
	for m.flag[other] == 1 && m.turn == int32(other) {
		// spin
	}
}

func (m *PetersonMutex) Unlock(self int) {
	m.flag[self] = 0
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
	fmt.Printf("Итоговое значение счётчика: %d (ожидается 200, но скорее всего нет)\n", sharedCounter)
}

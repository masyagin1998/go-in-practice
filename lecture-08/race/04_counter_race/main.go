package main

// Data race на счётчике: counter++ — это RMW (read → +1 → write).
// Без синхронизации две goroutine теряют часть инкрементов.

import (
	"fmt"
	"sync"
)

func main() {
	var counter int
	var wg sync.WaitGroup

	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1_000_000 {
				counter++
			}
		}()
	}

	wg.Wait()
	fmt.Printf("counter = %d (ожидалось 2000000)\n", counter)
}

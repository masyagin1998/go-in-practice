// go run main.go
//
// Buffered channel — канал с буфером.
// Send блокируется только когда буфер полон.
// Receive блокируется только когда буфер пуст.
// Позволяет отправителю и получателю работать с разной скоростью.

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// --- 1. Базовый буферизованный канал ---
	ch := make(chan int, 3) // буфер на 3 элемента

	ch <- 1 // не блокируется
	ch <- 2 // не блокируется
	ch <- 3 // не блокируется
	// ch <- 4 // заблокировалось бы! буфер полон

	fmt.Println(<-ch, <-ch, <-ch)

	// --- 2. Producer-consumer с разной скоростью ---
	fmt.Println("\n=== Producer-Consumer ===")
	jobs := make(chan int, 5) // буфер сглаживает разницу скоростей

	// Producer — быстрый.
	go func() {
		for i := 0; i < 10; i++ {
			jobs <- i
			fmt.Printf("Produced: %d (len=%d)\n", i, len(jobs))
		}
		close(jobs)
	}()

	// Consumer — медленный.
	for job := range jobs {
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("Consumed: %d\n", job)
	}

	// --- 3. Semaphore — ограничение параллелизма ---
	fmt.Println("\n=== Semaphore ===")
	sem := make(chan struct{}, 3) // максимум 3 одновременно

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{} // захватить слот
			fmt.Printf("Worker %d: работаю (active=%d)\n", id, len(sem))
			time.Sleep(100 * time.Millisecond)
			<-sem // освободить слот
		}(i)
	}
	wg.Wait()
}

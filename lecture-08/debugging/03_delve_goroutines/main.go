package main

// Четыре воркера завершаются, пятая goroutine зависает на recv из канала,
// в который никто не пишет. Учимся переключаться между goroutine'ами в dlv.

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("[worker %d] закончил\n", id)
}

func hang(ch <-chan int) {
	v := <-ch
	fmt.Println("[hang] получил:", v)
}

func main() {
	stuck := make(chan int)

	var wg sync.WaitGroup
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}
	go hang(stuck)

	wg.Wait()
	fmt.Println("[main] воркеры отработали, но одна goroutine висит")
	time.Sleep(2 * time.Second) // даём время посмотреть в dlv
}

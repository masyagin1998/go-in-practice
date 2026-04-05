// go run main.go
//
// Fan-In: объединяем несколько каналов в один.
// Каждый источник генерирует данные с разной скоростью,
// результат мультиплексируется в единый выходной канал.

package main

import (
	"fmt"
	"sync"
	"time"
)

// fanIn объединяет произвольное число каналов в один.
func fanIn(channels ...<-chan string) <-chan string {
	out := make(chan string)

	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// source создаёт канал, который генерирует n сообщений с заданным интервалом.
func source(name string, interval time.Duration, n int) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for i := range n {
			ch <- fmt.Sprintf("%s: сообщение %d", name, i)
			time.Sleep(interval)
		}
	}()
	return ch
}

func main() {
	// Три источника с разной частотой
	fast := source("fast", 50*time.Millisecond, 5)
	medium := source("medium", 150*time.Millisecond, 3)
	slow := source("slow", 300*time.Millisecond, 2)

	// Объединяем в один поток
	merged := fanIn(fast, medium, slow)

	for msg := range merged {
		fmt.Println(msg)
	}
}

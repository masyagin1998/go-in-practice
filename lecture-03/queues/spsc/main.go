// go run main.go
//
// SPSC (Single Producer — Single Consumer):
// один producer отправляет сообщения в unbuffered канал,
// один consumer читает из него.

package main

import (
	"fmt"
	"sync"
)

func main() {
	ch := make(chan int)
	var wg sync.WaitGroup

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 5 {
			ch <- i
			fmt.Printf("[producer] отправил %d\n", i)
		}
		close(ch)
	}()

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ch {
			fmt.Printf("[consumer] получил %d\n", v)
		}
	}()

	wg.Wait()
}

// go run main.go
//
// SPMC (Single Producer — Multiple Consumers):
// один producer отправляет задачи в unbuffered канал,
// несколько consumer'ов забирают и обрабатывают их.

package main

import (
	"fmt"
	"sync"
)

func main() {
	ch := make(chan int)
	var wg sync.WaitGroup

	// Запускаем 3 consumer'а
	for c := range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				fmt.Printf("[consumer-%d] обработал %d\n", c, v)
			}
		}()
	}

	// Producer (в main-горутине)
	for i := range 10 {
		ch <- i
	}
	close(ch)

	wg.Wait()
}

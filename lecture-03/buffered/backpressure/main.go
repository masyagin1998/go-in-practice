// go run main.go
//
// Backpressure: быстрый producer, медленный consumer.
// Буфер канала сглаживает разницу в скорости, но когда буфер
// заполняется — producer блокируется, давая consumer время догнать.

package main

import (
	"fmt"
	"time"
)

func main() {
	// Буфер на 3 элемента — producer может уйти вперёд максимум на 3
	ch := make(chan int, 3)

	// Быстрый producer: генерирует элемент каждые 50ms
	go func() {
		for i := range 10 {
			start := time.Now()
			ch <- i
			wait := time.Since(start)
			fmt.Printf("[producer] отправил %d (ждал %v)\n", i, wait.Round(time.Millisecond))
		}
		close(ch)
	}()

	// Медленный consumer: обрабатывает элемент 200ms
	for v := range ch {
		fmt.Printf("[consumer] получил %d\n", v)
		time.Sleep(200 * time.Millisecond)
	}
}

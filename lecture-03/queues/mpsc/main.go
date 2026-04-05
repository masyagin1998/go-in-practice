// go run main.go
//
// MPSC (Multiple Producers — Single Consumer):
// несколько producer'ов пишут в один unbuffered канал,
// один consumer читает из него.

package main

import (
	"fmt"
	"sync"
)

func main() {
	ch := make(chan string)
	var wg sync.WaitGroup

	// Запускаем 3 producer'а
	for p := range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 3 {
				msg := fmt.Sprintf("producer-%d: сообщение %d", p, i)
				ch <- msg
			}
		}()
	}

	// Закрываем канал после завершения всех producer'ов
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Consumer (в main-горутине)
	for msg := range ch {
		fmt.Printf("[consumer] %s\n", msg)
	}
}

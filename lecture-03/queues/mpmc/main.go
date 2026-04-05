// go run main.go
//
// MPMC (Multiple Producers — Multiple Consumers):
// несколько producer'ов пишут задачи в unbuffered канал,
// несколько consumer'ов читают и обрабатывают их.

package main

import (
	"fmt"
	"sync"
)

func main() {
	ch := make(chan string)

	var producers sync.WaitGroup
	var consumers sync.WaitGroup

	// Запускаем 3 producer'а
	for p := range 3 {
		producers.Add(1)
		go func() {
			defer producers.Done()
			for i := range 3 {
				msg := fmt.Sprintf("producer-%d: задача %d", p, i)
				ch <- msg
			}
		}()
	}

	// Запускаем 2 consumer'а
	for c := range 2 {
		consumers.Add(1)
		go func() {
			defer consumers.Done()
			for msg := range ch {
				fmt.Printf("[consumer-%d] %s\n", c, msg)
			}
		}()
	}

	// Закрываем канал после завершения всех producer'ов
	producers.Wait()
	close(ch)

	// Ждём завершения всех consumer'ов
	consumers.Wait()
}

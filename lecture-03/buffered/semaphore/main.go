// go run main.go
//
// Семафор через buffered channel: ограничиваем число одновременно
// выполняемых задач. Буфер канала = максимальное число параллельных задач.
// Горутина захватывает слот записью в канал, освобождает — чтением.

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	const maxConcurrent = 3
	// Буфер = максимум одновременно работающих горутин
	sem := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			sem <- struct{}{} // захватываем слот
			defer func() { <-sem }() // освобождаем слот

			fmt.Printf("[задача %d] начало\n", i)
			time.Sleep(300 * time.Millisecond) // имитация работы
			fmt.Printf("[задача %d] конец\n", i)
		}()
	}

	wg.Wait()
}

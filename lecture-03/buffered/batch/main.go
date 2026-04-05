// go run main.go
//
// Батчинг: накапливаем элементы в буфере и обрабатываем пачками.
// Батч отправляется когда набралось batchSize элементов ИЛИ
// сработал таймаут — что произойдёт раньше.

package main

import (
	"fmt"
	"time"
)

func main() {
	const batchSize = 4

	ch := make(chan int, batchSize)

	// Producer: генерирует события с переменной частотой
	go func() {
		delays := []int{10, 10, 10, 10, 500, 10, 10, 500, 10, 10, 10, 10}
		for i, d := range delays {
			ch <- i
			time.Sleep(time.Duration(d) * time.Millisecond)
		}
		close(ch)
	}()

	// Consumer: собирает батчи
	batch := make([]int, 0, batchSize)
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case v, ok := <-ch:
			if !ok {
				// Канал закрыт — обрабатываем остаток
				if len(batch) > 0 {
					fmt.Printf("[batch] финальный: %v (размер %d)\n", batch, len(batch))
				}
				return
			}
			batch = append(batch, v)
			if len(batch) >= batchSize {
				fmt.Printf("[batch] по размеру: %v\n", batch)
				batch = batch[:0]
				timer.Reset(200 * time.Millisecond)
			}

		case <-timer.C:
			// Таймаут — отправляем что накопилось
			if len(batch) > 0 {
				fmt.Printf("[batch] по таймауту: %v (размер %d)\n", batch, len(batch))
				batch = batch[:0]
			}
			timer.Reset(200 * time.Millisecond)
		}
	}
}

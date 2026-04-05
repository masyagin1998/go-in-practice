// go run main.go
//
// Or-Done: безопасное чтение из канала с возможностью отмены.
// Оборачиваем любой канал так, чтобы при закрытии done-канала
// чтение немедленно прекращалось — даже если источник ещё жив.

package main

import (
	"fmt"
	"time"
)

// orDone оборачивает канал: читает из in, пока не закроется done.
func orDone(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					return
				}
			}
		}
	}()
	return out
}

func main() {
	// Бесконечный источник
	infinite := make(chan int)
	go func() {
		i := 0
		for {
			infinite <- i
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Через 550ms отменяем чтение
	done := make(chan struct{})
	go func() {
		time.Sleep(550 * time.Millisecond)
		close(done)
		fmt.Println("[done] отмена!")
	}()

	for v := range orDone(done, infinite) {
		fmt.Printf("получили: %d\n", v)
	}

	fmt.Println("завершено")
}

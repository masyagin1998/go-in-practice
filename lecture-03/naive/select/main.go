// go run main.go
//
// select ждёт первый готовый канал из нескольких.
// default делает select неблокирующим.

package main

import (
	"fmt"
	"time"
)

func main() {
	// --- 1. Обычный select: ждём первый готовый канал ---
	fmt.Println("=== select ===")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch1 <- "медленный"
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch2 <- "быстрый"
	}()

	select {
	case msg := <-ch1:
		fmt.Println("ch1:", msg)
	case msg := <-ch2:
		fmt.Println("ch2:", msg)
	}

	// --- 2. select + default: неблокирующая проверка канала ---
	fmt.Println("\n=== select + default ===")
	ch := make(chan int, 1)

	// Канал пуст — сработает default.
	select {
	case v := <-ch:
		fmt.Println("получили:", v)
	default:
		fmt.Println("канал пуст, идём дальше")
	}

	// Положим значение и проверим ещё раз.
	ch <- 42

	select {
	case v := <-ch:
		fmt.Println("получили:", v)
	default:
		fmt.Println("канал пуст, идём дальше")
	}
}

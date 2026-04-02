// go run main.go
//
// select — мультиплексирование каналов.
// Ждёт первый готовый канал. Если готовы несколько — выбирает случайный.

package main

import (
	"fmt"
	"time"
)

func main() {
	basicSelect()
	fmt.Println()
	timeoutExample()
	fmt.Println()
	nonBlockingExample()
	fmt.Println()
	fanInExample()
}

// --- 1. Базовый select: кто первый ---
func basicSelect() {
	fmt.Println("=== Basic Select ===")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "one"
	}()
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch2 <- "two"
	}()

	select {
	case msg := <-ch1:
		fmt.Println("Получили из ch1:", msg)
	case msg := <-ch2:
		fmt.Println("Получили из ch2:", msg) // ch2 быстрее
	}
}

// --- 2. Timeout через select ---
func timeoutExample() {
	fmt.Println("=== Timeout ===")
	ch := make(chan string)

	go func() {
		time.Sleep(500 * time.Millisecond) // слишком долго
		ch <- "result"
	}()

	select {
	case msg := <-ch:
		fmt.Println("Получили:", msg)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Таймаут! Не дождались ответа.")
	}
}

// --- 3. Non-blocking операция через default ---
func nonBlockingExample() {
	fmt.Println("=== Non-blocking ===")
	ch := make(chan int, 1)

	// Non-blocking send.
	select {
	case ch <- 42:
		fmt.Println("Отправили 42")
	default:
		fmt.Println("Канал полон, пропускаем")
	}

	// Non-blocking receive.
	select {
	case v := <-ch:
		fmt.Printf("Получили: %d\n", v)
	default:
		fmt.Println("Канал пуст")
	}

	// Ещё раз — канал уже пуст.
	select {
	case v := <-ch:
		fmt.Printf("Получили: %d\n", v)
	default:
		fmt.Println("Канал пуст")
	}
}

// --- 4. Fan-in: объединение нескольких каналов в один ---
func fanInExample() {
	fmt.Println("=== Fan-in ===")
	ch1 := produce("A", 50*time.Millisecond)
	ch2 := produce("B", 80*time.Millisecond)

	// Читаем из обоих каналов 10 сообщений.
	for range 10 {
		select {
		case msg := <-ch1:
			fmt.Println(msg)
		case msg := <-ch2:
			fmt.Println(msg)
		}
	}
}

func produce(name string, interval time.Duration) <-chan string {
	ch := make(chan string)
	go func() {
		for i := 0; ; i++ {
			time.Sleep(interval)
			ch <- fmt.Sprintf("%s: сообщение %d", name, i)
		}
	}()
	return ch
}

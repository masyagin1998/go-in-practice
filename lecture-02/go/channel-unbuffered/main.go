// go run main.go
//
// Unbuffered channel — канал без буфера.
// Send блокируется, пока кто-то не сделает Receive (и наоборот).
// Это точка синхронизации: отправитель и получатель встречаются.

package main

import (
	"fmt"
	"time"
)

func main() {
	// --- 1. Простая передача значения ---
	ch := make(chan string) // unbuffered

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- "hello" // блокируется, пока main не прочитает
		fmt.Println("Sender: отправил")
	}()

	msg := <-ch // блокируется, пока sender не отправит
	fmt.Printf("Receiver: получил %q\n", msg)

	// --- 2. Пинг-понг — синхронная передача между горутинами ---
	ping := make(chan int)
	pong := make(chan int)

	go func() {
		for v := range ping {
			fmt.Printf("  ping: получил %d, отправляю %d\n", v, v+1)
			pong <- v + 1
		}
	}()

	for i := 0; i < 5; i++ {
		ping <- i
		result := <-pong
		fmt.Printf("  main: отправил %d, получил %d\n", i, result)
	}
	close(ping)

	// --- 3. Done-канал — сигнал завершения ---
	done := make(chan struct{})

	go func() {
		fmt.Println("\nWorker: работаю...")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("Worker: готово")
		close(done) // сигнал: работа завершена (close разблокирует всех читателей)
	}()

	<-done // ждём
	fmt.Println("Main: worker завершился")
}

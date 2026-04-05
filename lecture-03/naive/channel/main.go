// go run main.go
//
// Простейший unbuffered канал: горутина отправляет, main получает.

package main

import "fmt"

func main() {
	ch := make(chan string)

	go func() {
		ch <- "привет из горутины"
	}()

	msg := <-ch
	fmt.Println(msg)
}

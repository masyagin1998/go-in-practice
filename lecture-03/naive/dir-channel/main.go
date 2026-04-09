package main

import "fmt"

// функция только читает из канала
func reader(ch <-chan string) {
	msg := <-ch
	ch <- "hello from writer"
	fmt.Println("reader got:", msg)
}

// функция только пишет в канал
func writer(ch chan<- string) {
	ch <- "hello from writer"
}

func main() {
	ch := make(chan string)

	go writer(ch) // передаём как write-only
	reader(ch)    // передаём как read-only
}

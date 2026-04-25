package main

// Утечка goroutine на канале — самый частый паттерн.
// Send в unbuffered канал без читателя или recv из канала, в который
// никто не пишет, блокируют goroutine навсегда. Канал и goroutine
// держат друг друга → GC бессилен.

import (
	"fmt"
	"runtime"
	"time"
)

func leakOnSend() {
	ch := make(chan int)
	go func() {
		ch <- 42 // никто не читает, висим навсегда
	}()
}

func leakOnRecv() {
	ch := make(chan int)
	go func() {
		<-ch // никто не пишет, висим навсегда
	}()
}

func main() {
	before := runtime.NumGoroutine()

	for range 5 {
		leakOnSend()
	}
	for range 5 {
		leakOnRecv()
	}

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()

	fmt.Printf("до:     %d\n", before)
	fmt.Printf("после:  %d (+%d)\n", after, after-before)
}

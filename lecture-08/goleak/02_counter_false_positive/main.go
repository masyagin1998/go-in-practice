package main

// NumGoroutine() — это снимок счётчика в моменте, а не отчёт об утечках.
// Демонстрируем два классических false-positive: "ещё не успели выйти"
// и "это вообще не наши goroutine".

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// slowShutdown: честная goroutine, выходит сама через 200 мс.
func slowShutdown() {
	go func() {
		time.Sleep(200 * time.Millisecond)
	}()
}

func main() {
	before := runtime.NumGoroutine()

	// Кейс 1. Ещё не успели выйти.
	for range 10 {
		slowShutdown()
	}
	time.Sleep(20 * time.Millisecond)
	early := runtime.NumGoroutine()
	fmt.Printf("ранний снимок (через 20мс):  +%d goroutine\n", early-before)
	fmt.Println("→ выглядит как 10 утёкших — но они просто не успели завершиться")

	time.Sleep(300 * time.Millisecond)
	late := runtime.NumGoroutine()
	fmt.Printf("поздний снимок (через 320мс): +%d goroutine\n", late-before)
	fmt.Println("→ утечки нет, все вышли сами")

	// Кейс 2. Чужие goroutine. http.Get при первом обращении к хосту
	// поднимает фоновые readLoop / writeLoop для keep-alive — они
	// "вечные" и в нашем коде про них ничего нет.
	fmt.Println()
	resp, err := http.Get("https://example.com/")
	if err == nil {
		_ = resp.Body.Close()
	}
	// Подождём долго — больше IdleConnTimeout по умолчанию это не сделает,
	// но точно дольше любого "ещё не успели". И всё равно увидим +N.
	time.Sleep(2 * time.Second)
	httpExtra := runtime.NumGoroutine() - late
	fmt.Printf("после http.Get + 2с ожидания: +%d goroutine\n", httpExtra)
	fmt.Println("→ это keep-alive readLoop/writeLoop из net/http; ждать бесполезно")
}

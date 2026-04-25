package main

// runtime/trace: видим, как планировщик и GC ведут себя на живой
// рабочей нагрузке.
//
// Что делаем:
//   1. Пул из 8 goroutine шлёт друг другу сообщения через каналы.
//   2. Параллельно — "аллокатор-монстр", чтобы GC честно работал.
//   3. Всё это обёрнуто в runtime/trace.Start/Stop.
//
// Смотрим:
//   go tool trace trace.out
//   → вкладки "Goroutine analysis", "Scheduler latency profile",
//     "Network blocking profile", "Syscall blocking profile", "GC".
//   → "View trace" — таймлайн всех P/M/G.

import (
	"fmt"
	"os"
	"runtime/trace"
	"sync"
)

func pingPong(n int, from, to chan int) {
	for i := 0; i < n; i++ {
		v := <-from
		to <- v + 1
	}
}

func allocator(iters int) {
	for i := 0; i < iters; i++ {
		// Нагружаем GC: часто выделяем и отпускаем.
		b := make([]byte, 16*1024)
		_ = b
	}
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()

	// ping-pong между парами goroutine
	const pairs = 4
	var wg sync.WaitGroup
	chans := make([]chan int, pairs*2)
	for i := range chans {
		chans[i] = make(chan int, 1)
	}
	for p := 0; p < pairs; p++ {
		a := chans[p*2]
		b := chans[p*2+1]
		wg.Add(2)
		go func() { defer wg.Done(); pingPong(500, a, b) }()
		go func() { defer wg.Done(); pingPong(500, b, a) }()
		a <- 0 // стартовый мяч
	}

	// параллельно — аллокатор
	wg.Add(1)
	go func() { defer wg.Done(); allocator(20_000) }()

	wg.Wait()
	fmt.Println("trace.out готов; открывай: go tool trace trace.out")
}

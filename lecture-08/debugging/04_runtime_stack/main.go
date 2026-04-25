package main

// Дамп стеков всех goroutine без дебагера: SIGUSR1 → runtime.Stack →
// stderr, процесс продолжает работу. Альтернатива без кода:
// GOTRACEBACK=all + SIGQUIT (процесс умирает, но не нужна перекомпиляция).

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func worker(id int, wg *sync.WaitGroup, done <-chan struct{}) {
	defer wg.Done()
	fmt.Printf("[worker %d] запущен\n", id)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			fmt.Printf("[worker %d] выходит\n", id)
			return
		case <-ticker.C:
		}
	}
}

func dumpStacks() {
	buf := make([]byte, 1<<14)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			fmt.Fprintf(os.Stderr, "=== stacks (%d goroutine) ===\n%s\n",
				runtime.NumGoroutine(), buf[:n])
			return
		}
		buf = make([]byte, 2*len(buf))
	}
}

func main() {
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(i, &wg, done)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("[main] PID=%d, kill -USR1 %d для дампа, Ctrl-C для выхода\n",
		os.Getpid(), os.Getpid())

	for {
		s := <-sigs
		if s == syscall.SIGUSR1 {
			dumpStacks()
			continue
		}
		fmt.Println("[main] выходим по", s)
		close(done)
		wg.Wait()
		return
	}
}

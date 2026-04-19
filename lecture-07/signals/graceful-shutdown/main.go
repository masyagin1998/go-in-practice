package main

// graceful-shutdown — SIGINT/SIGTERM → ctx.Cancel → воркеры доделывают
// текущий item и выходят. Каркас большинства Go-сервисов.

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const workers = 3

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	work := make(chan int, 16)
	var wg sync.WaitGroup
	for id := 1; id <= workers; id++ {
		wg.Add(1)
		go worker(ctx, id, work, &wg)
	}

	// Продюсер: кидаем задачи пока контекст жив.
	go func() {
		defer close(work)
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return
			case work <- i:
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	fmt.Printf("PID=%d, Ctrl+C или kill -TERM %d для graceful shutdown\n",
		os.Getpid(), os.Getpid())

	<-ctx.Done()
	fmt.Println("\n[main] сигнал получен, ждём воркеров...")
	wg.Wait()
	fmt.Println("[main] все воркеры завершились")
}

func worker(ctx context.Context, id int, in <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			// Долгий item (150ms). Внутри тоже проверяем ctx — иначе
			// самая долгая задача удерживает shutdown.
			select {
			case <-ctx.Done():
				fmt.Printf("[w%d] прерываем item %d\n", id, item)
				return
			case <-time.After(150 * time.Millisecond):
				fmt.Printf("[w%d] item %d\n", id, item)
			}
		}
	}
}

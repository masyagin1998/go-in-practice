package worker

import (
	"context"
	"sync"
	"time"
)

// Clean — правильная: goroutine выходит по ctx.Done.
func Clean(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(50 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
}

// Leaky — та же сигнатура, ctx игнорируется. Снаружи код передаёт ctx
// и зовёт cancel() — кажется правильным. А goroutine крутит ticker
// без оглядки на контекст и живёт вечно.
func Leaky(_ context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		// defer wg.Done() намеренно нет.
		t := time.NewTicker(50 * time.Millisecond)
		defer t.Stop()
		for range t.C {
		}
	}()
}

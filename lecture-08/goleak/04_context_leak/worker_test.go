package worker

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestWorkerLeaky(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int)
	out := make(chan int, 1)

	WorkerLeaky(ctx, in, out)

	cancel()
	time.Sleep(50 * time.Millisecond)
	// goleak увидит висящего consumer'а на <-in.
}

func TestWorkerFixed(t *testing.T) {
	// IgnoreCurrent: goroutine из TestWorkerLeaky всё ещё жива в этом
	// процессе. Снимаем baseline "что уже висит" — будем сравнивать
	// только с новыми.
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int)
	out := make(chan int, 1)

	WorkerFixed(ctx, in, out)

	cancel()
	time.Sleep(50 * time.Millisecond)
}

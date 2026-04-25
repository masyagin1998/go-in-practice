package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// Способ интеграции goleak — VerifyNone в каждом тесте через defer.
// Альтернатива: TestMain(m) + goleak.VerifyTestMain(m) — одна проверка
// на весь пакет, но тогда любая утечка красит весь пакет.

func TestClean(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	Clean(ctx, &wg)

	time.Sleep(120 * time.Millisecond)
	cancel()
	wg.Wait()
}

func TestLeaky(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Setup/teardown такие же, как в TestClean — снаружи всё выглядит
	// правильно. goleak ловит утечку по факту, а не по сигнатуре.
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	Leaky(ctx, &wg)

	time.Sleep(120 * time.Millisecond)
	cancel()

	// wg.Wait() здесь нельзя: Leaky не зовёт wg.Done() → зависнем.
}

// Утечка контекста: context.WithCancel / WithTimeout создают внутренние
// горутины и структуры. Если не вызвать cancel() — они не освободятся
// до завершения родительского контекста.
//
// Сборка:
//   go build -o 05_context_leak .
//
// Запуск:
//   ./05_context_leak
//
// Запуск с valgrind:
//   valgrind --tool=memcheck ./05_context_leak

package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// processLeaky создаёт контекст с таймаутом, но не вызывает cancel.
func processLeaky(parent context.Context) {
	// WithTimeout создаёт внутреннюю горутину для отслеживания таймера.
	ctx, _ := context.WithTimeout(parent, 5*time.Second)
	// Забыли: defer cancel()

	// Используем контекст для какой-то быстрой операции.
	select {
	case <-time.After(1 * time.Millisecond):
		// Операция завершена быстро, но таймер на 5 секунд всё ещё тикает.
		// Внутренняя горутина контекста будет жить ещё ~5 секунд.
	case <-ctx.Done():
	}
}

// processCorrect — правильный вариант.
func processCorrect(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel() // Немедленно освобождаем ресурсы контекста.

	select {
	case <-time.After(1 * time.Millisecond):
	case <-ctx.Done():
	}
}

func main() {
	parent := context.Background()

	fmt.Printf("Горутин в начале: %d\n", runtime.NumGoroutine())

	// Создаём 100 «утекающих» контекстов.
	for i := 0; i < 100; i++ {
		processLeaky(parent)
	}
	fmt.Printf("Горутин после утечки контекстов: %d\n", runtime.NumGoroutine())

	// Ждём, пока таймауты истекут.
	time.Sleep(6 * time.Second)
	fmt.Printf("Горутин после истечения таймаутов: %d\n", runtime.NumGoroutine())

	// Правильный вариант: горутины освобождаются сразу.
	for i := 0; i < 100; i++ {
		processCorrect(parent)
	}
	fmt.Printf("Горутин после корректных контекстов: %d\n", runtime.NumGoroutine())
}

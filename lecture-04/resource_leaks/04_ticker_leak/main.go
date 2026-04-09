// Утечка тикера: time.NewTicker создаёт горутину внутри рантайма,
// которая периодически отправляет значения в канал.
// Если не вызвать Stop() — горутина и канал останутся навсегда.
//
// Сборка:
//   go build -o 04_ticker_leak .
//
// Запуск:
//   ./04_ticker_leak
//
// Запуск с valgrind:
//   valgrind --tool=memcheck ./04_ticker_leak

package main

import (
	"fmt"
	"runtime"
	"time"
)

// monitorLeaky запускает мониторинг, но не останавливает тикер.
func monitorLeaky(name string, duration time.Duration) {
	ticker := time.NewTicker(10 * time.Millisecond)
	// Забыли: defer ticker.Stop()

	go func() {
		timeout := time.After(duration)
		for {
			select {
			case t := <-ticker.C:
				_ = t // Обрабатываем тик.
			case <-timeout:
				fmt.Printf("  [%s] Мониторинг завершён, но тикер не остановлен!\n", name)
				return
				// Горутина завершилась, но тикер продолжает тикать в пустоту.
				// Внутренняя горутина тикера — утечка.
			}
		}
	}()
}

// monitorCorrect — правильный вариант с ticker.Stop().
func monitorCorrect(name string, duration time.Duration) {
	ticker := time.NewTicker(10 * time.Millisecond)

	go func() {
		defer ticker.Stop() // Останавливаем тикер при выходе.
		timeout := time.After(duration)
		for {
			select {
			case t := <-ticker.C:
				_ = t
			case <-timeout:
				fmt.Printf("  [%s] Мониторинг завершён, тикер остановлен.\n", name)
				return
			}
		}
	}()
}

func main() {
	fmt.Printf("Горутин в начале: %d\n", runtime.NumGoroutine())

	// Запускаем 10 «утекающих» мониторов.
	for i := 0; i < 10; i++ {
		monitorLeaky(fmt.Sprintf("leaky-%d", i), 50*time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Горутин после утечки тикеров: %d\n", runtime.NumGoroutine())

	// Запускаем 10 корректных мониторов.
	for i := 0; i < 10; i++ {
		monitorCorrect(fmt.Sprintf("correct-%d", i), 50*time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Горутин после корректных тикеров: %d\n", runtime.NumGoroutine())
}

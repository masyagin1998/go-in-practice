// go run main.go
//
// sync/atomic — атомарные операции без мьютексов.
// Быстрее Mutex для простых операций (счётчики, флаги).

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// --- 1. atomic.Int64 — типизированный счётчик (Go 1.19+) ---
	var counter atomic.Int64
	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1)
		}()
	}
	wg.Wait()
	fmt.Printf("atomic.Int64: %d (ожидается 1000)\n", counter.Load())

	// --- 2. atomic.Bool — потокобезопасный флаг ---
	var shutdown atomic.Bool

	go func() {
		// Воркер проверяет флаг.
		for !shutdown.Load() {
			// работаем...
		}
		fmt.Println("Воркер остановлен.")
	}()

	shutdown.Store(true)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Ждём чтобы воркер успел увидеть флаг.
	}()
	wg.Wait()

	// --- 3. CompareAndSwap — основа lock-free алгоритмов ---
	var state atomic.Int32
	state.Store(0) // idle

	// Только одна горутина "захватит" переход 0→1.
	winners := atomic.Int32{}
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if state.CompareAndSwap(0, 1) {
				winners.Add(1)
				state.Store(0) // отпускаем для следующего
			}
		}()
	}
	wg.Wait()
	fmt.Printf("CAS: %d горутин смогли захватить переход 0→1\n", winners.Load())

	// --- 4. atomic.Value — хранение произвольного значения ---
	var config atomic.Value
	config.Store(map[string]string{"env": "prod", "port": "8080"})

	// Читатели получают snapshot — никаких мьютексов.
	cfg := config.Load().(map[string]string)
	fmt.Printf("atomic.Value: env=%s, port=%s\n", cfg["env"], cfg["port"])

	// Обновление — атомарная замена целиком.
	config.Store(map[string]string{"env": "staging", "port": "9090"})
	cfg = config.Load().(map[string]string)
	fmt.Printf("atomic.Value (updated): env=%s, port=%s\n", cfg["env"], cfg["port"])
}

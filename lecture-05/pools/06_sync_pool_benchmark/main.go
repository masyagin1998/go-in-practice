// Пример 6: Бенчмарк — аллокация vs sync.Pool vs channel pool.
//
// Сравниваем три стратегии:
//   1. Прямая аллокация (make каждый раз)
//   2. sync.Pool
//   3. Channel pool (буферизованный канал)
//
// Измеряем: время, число аллокаций, давление на GC.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	iterations = 1_000_000
	bufSize    = 4096
	poolChSize = 256 // размер канала для channel pool
)

// sink предотвращает оптимизацию компилятором.
var sink []byte

// benchAlloc — каждый раз make.
func benchAlloc(n int) {
	for range n {
		buf := make([]byte, 0, bufSize)
		buf = append(buf, 0xAB)
		sink = buf
	}
}

// benchSyncPool — sync.Pool.
func benchSyncPool(n int) {
	pool := sync.Pool{
		New: func() any {
			return make([]byte, 0, bufSize)
		},
	}
	for range n {
		buf := pool.Get().([]byte)
		buf = buf[:0]
		buf = append(buf, 0xAB)
		pool.Put(buf)
	}
}

// benchChannelPool — buffered channel.
func benchChannelPool(n int) {
	ch := make(chan []byte, poolChSize)
	// Прогреваем канал.
	for range poolChSize {
		ch <- make([]byte, 0, bufSize)
	}

	for range n {
		buf := <-ch
		buf = buf[:0]
		buf = append(buf, 0xAB)
		ch <- buf
	}
}

// measure — замеряет время и статистику аллокаций.
func measure(name string, fn func(int), n int) {
	runtime.GC()
	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()
	fn(n)
	dur := time.Since(start)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	mallocs := after.Mallocs - before.Mallocs
	totalAlloc := after.TotalAlloc - before.TotalAlloc
	gcCycles := after.NumGC - before.NumGC

	fmt.Printf("  %-22s %8v  mallocs: %-10d  bytes: %-12d  GC: %d\n",
		name, dur.Round(time.Millisecond), mallocs, totalAlloc, gcCycles)
}

func main() {
	fmt.Printf("=== Бенчмарк пулов (%d итераций, буфер %d байт) ===\n\n",
		iterations, bufSize)

	// Однопоточный тест.
	fmt.Println("--- Однопоточный ---")
	measure("make (каждый раз)", benchAlloc, iterations)
	measure("sync.Pool", benchSyncPool, iterations)
	measure("channel pool", benchChannelPool, iterations)

	// Многопоточный тест.
	fmt.Printf("\n--- Многопоточный (GOMAXPROCS=%d) ---\n", runtime.GOMAXPROCS(0))

	concurrentBench := func(name string, fn func(int)) {
		workers := runtime.GOMAXPROCS(0)
		perWorker := iterations / workers

		measure(name, func(_ int) {
			var wg sync.WaitGroup
			for range workers {
				wg.Go(func() {
					fn(perWorker)
				})
			}
			wg.Wait()
		}, iterations)
	}

	concurrentBench("make (каждый раз)", benchAlloc)
	concurrentBench("sync.Pool", benchSyncPool)
	// Channel pool в многопоточном тесте показывает contention на канале.
	concurrentBench("channel pool", func(n int) {
		ch := make(chan []byte, poolChSize)
		for range poolChSize {
			ch <- make([]byte, 0, bufSize)
		}
		for range n {
			buf := <-ch
			buf = buf[:0]
			buf = append(buf, 0xAB)
			ch <- buf
		}
	})

	fmt.Println()
	fmt.Println("  → sync.Pool выигрывает в многопоточном режиме:")
	fmt.Println("    у каждого P есть локальный кэш (private + shared),")
	fmt.Println("    нет contention как у общего канала или мьютекса.")
	fmt.Println()
	fmt.Println("  → Channel pool: contention растёт с числом горутин.")
	fmt.Println("    Каждый send/recv — это lock на внутреннем мьютексе канала.")
	fmt.Println()
	fmt.Println("  → make каждый раз: давление на GC, но зато нет contention")
	fmt.Println("    и нет грязных данных.")
}

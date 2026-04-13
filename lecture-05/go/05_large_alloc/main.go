// Пример 5: Большие аллокации в Go (>32 КБ).
//
// Объекты <= 32 КБ: mcache → mcentral → mheap (size-классы, span'ы).
// Объекты > 32 КБ: напрямую в mheap, размер округляется до страниц (8 КБ).
//
// Это аналог того, как glibc malloc использует mmap для больших аллокаций
// вместо brk — обходит пул мелких объектов.
//
// Демонстрируем:
//   1. Порог 32 КБ — скачок в HeapInuse при переходе через границу.
//   2. Округление до страниц для больших объектов.
//   3. Разницу в скорости аллокации маленьких vs больших объектов.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

func measure(label string, n int, allocFn func()) {
	runtime.GC()
	debug.FreeOSMemory()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()
	for i := 0; i < n; i++ {
		allocFn()
	}
	elapsed := time.Since(start)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	fmt.Printf("  %-40s %6d allocs, %v, +%d КБ heap\n",
		label, n, elapsed,
		(after.HeapInuse-before.HeapInuse)/1024)
}

func main() {
	// =============================================================
	// Тест 1: Порог 32 КБ — small vs large path
	// =============================================================
	fmt.Println("=== Тест 1: Порог 32 КБ ===")
	fmt.Println("  Объекты ровно на границе: 32768 (small) vs 32769 (large)")
	fmt.Println()

	const N = 1000

	// 32768 = 32 КБ — последний size-класс (#67)
	measure("[32768 байт — small path]", N, func() {
		p := make([]byte, 32768)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	// 32769 — уже large, идёт в mheap напрямую
	measure("[32769 байт — large path]", N, func() {
		p := make([]byte, 32769)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	// =============================================================
	// Тест 2: Округление до страниц (8 КБ)
	// =============================================================
	fmt.Println("\n=== Тест 2: Округление больших объектов до страниц (8 КБ) ===")

	sizes := []int{
		33 * 1024,  // 33 КБ → 5 стр (40 КБ), потери 17%
		40 * 1024,  // 40 КБ → 5 стр (40 КБ), потери 0%
		63 * 1024,  // 63 КБ → 8 стр (64 КБ), потери 1.6%
		64 * 1024,  // 64 КБ → 8 стр (64 КБ), потери 0%
		65 * 1024,  // 65 КБ → 9 стр (72 КБ), потери 9.7%
		100 * 1024, // 100 КБ → 13 стр (104 КБ), потери 3.8%
		1024 * 1024, // 1 МБ → 128 стр, потери 0%
	}

	fmt.Printf("  %-15s %-12s %-12s %-10s\n",
		"Запрошено", "Страниц", "Реально", "Потери")
	fmt.Println("  ---------------------------------------------------")

	pageSize := 8192
	for _, sz := range sizes {
		pages := (sz + pageSize - 1) / pageSize
		real := pages * pageSize
		waste := real - sz
		fmt.Printf("  %-15s %-12d %-12s %-10s\n",
			fmt.Sprintf("%d КБ", sz/1024),
			pages,
			fmt.Sprintf("%d КБ", real/1024),
			fmt.Sprintf("%d КБ (%.1f%%)", waste/1024, float64(waste)/float64(real)*100))
	}

	// =============================================================
	// Тест 3: Скорость аллокации — small vs large
	// =============================================================
	fmt.Println("\n=== Тест 3: Скорость аллокации ===")
	fmt.Println("  Small: mcache → быстрый путь, без лока")
	fmt.Println("  Large: mheap → глобальный лок")
	fmt.Println()

	// Маленькие
	measure("[1 КБ × 10000 — small path]", 10000, func() {
		p := make([]byte, 1024)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	measure("[8 КБ × 10000 — small path]", 10000, func() {
		p := make([]byte, 8192)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	measure("[32 КБ × 10000 — small path (max)]", 10000, func() {
		p := make([]byte, 32768)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	// Большие
	measure("[64 КБ × 10000 — large path]", 10000, func() {
		p := make([]byte, 64*1024)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	measure("[256 КБ × 10000 — large path]", 10000, func() {
		p := make([]byte, 256*1024)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	measure("[1 МБ × 1000 — large path]", 1000, func() {
		p := make([]byte, 1024*1024)
		p[0] = 1
		runtime.KeepAlive(p)
	})

	fmt.Println("\n  → Small аллокации значительно быстрее за счёт mcache (per-P, без лока).")
	fmt.Println("  → Large аллокации дороже: mheap лок + больше работы для GC.")
}

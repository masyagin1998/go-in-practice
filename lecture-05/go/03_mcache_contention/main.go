// Пример 3: Contention в аллокаторе Go при многопоточных аллокациях.
//
// Архитектура аллокатора Go:
//   mcache (per-P, без лока) → mcentral (per-size-class, спинлок) → mheap (глобальный лок)
//
// Быстрый путь: горутина аллоцирует из mcache своего P — без локов.
// Медленный путь: mcache пуст → идём в mcentral → лок.
//
// Contention возникает когда:
//   1. Много горутин на мало P — делят mcache.
//   2. Аллокации исчерпывают span в mcache → все идут в mcentral одновременно.
//   3. Большие объекты (>32 КБ) идут напрямую в mheap (глобальный лок).
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

// allocSmall — много мелких аллокаций (size-класс 32, попадает в mcache).
func allocSmall(n int) {
	for i := 0; i < n; i++ {
		p := new([32]byte)
		p[0] = byte(i) // чтобы компилятор не выкинул
		_ = p
	}
}

// allocLarge — аллокации >32 КБ (идут напрямую в mheap, глобальный лок).
func allocLarge(n int) {
	for i := 0; i < n; i++ {
		p := make([]byte, 64*1024) // 64 КБ — мимо mcache/mcentral
		p[0] = byte(i)
		_ = p
	}
}

// benchmark — запускает numG горутин, каждая делает allocFn(perG) аллокаций.
func benchmark(label string, numG, perG int, allocFn func(int)) time.Duration {
	// Прогрев
	allocFn(100)
	runtime.GC()

	var wg sync.WaitGroup
	wg.Add(numG)

	start := time.Now()
	for g := 0; g < numG; g++ {
		go func() {
			defer wg.Done()
			allocFn(perG)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("  %-45s %4d горутин × %-7d = %-10d  время: %v\n",
		label, numG, perG, numG*perG, elapsed)
	return elapsed
}

func main() {
	procs := runtime.GOMAXPROCS(0)
	fmt.Printf("GOMAXPROCS = %d\n\n", procs)

	const totalSmall = 2_000_000
	const totalLarge = 2_000

	// =============================================================
	// Тест 1: Мелкие аллокации (32 байта) — mcache fast path
	// =============================================================
	fmt.Println("=== Тест 1: Мелкие аллокации (32 байта, mcache fast path) ===")
	fmt.Println("  Ожидание: 1 горутина ≈ N горутин (mcache per-P, нет лока)")
	fmt.Println()

	t1 := benchmark("[1 горутина, все аллокации]",
		1, totalSmall, allocSmall)
	t2 := benchmark("[GOMAXPROCS горутин]",
		procs, totalSmall/procs, allocSmall)
	t3 := benchmark("[GOMAXPROCS×4 горутин]",
		procs*4, totalSmall/(procs*4), allocSmall)
	t4 := benchmark("[GOMAXPROCS×16 горутин]",
		procs*16, totalSmall/(procs*16), allocSmall)

	fmt.Println()
	fmt.Printf("  Ускорение %d горутин vs 1: %.2fx\n", procs, float64(t1)/float64(t2))
	fmt.Printf("  Ускорение %d горутин vs 1: %.2fx\n", procs*4, float64(t1)/float64(t3))
	fmt.Printf("  Ускорение %d горутин vs 1: %.2fx\n", procs*16, float64(t1)/float64(t4))

	// =============================================================
	// Тест 2: Большие аллокации (64 КБ) — mheap, глобальный лок
	// =============================================================
	fmt.Println("\n=== Тест 2: Большие аллокации (64 КБ, mheap global lock) ===")
	fmt.Println("  Ожидание: больше горутин → больше contention на mheap")
	fmt.Println()

	l1 := benchmark("[1 горутина, все аллокации]",
		1, totalLarge, allocLarge)
	l2 := benchmark("[GOMAXPROCS горутин]",
		procs, totalLarge/procs, allocLarge)
	l3 := benchmark("[GOMAXPROCS×4 горутин]",
		procs*4, totalLarge/(procs*4), allocLarge)
	l4 := benchmark("[GOMAXPROCS×16 горутин]",
		procs*16, totalLarge/(procs*16), allocLarge)

	fmt.Println()
	fmt.Printf("  Ускорение %d горутин vs 1: %.2fx\n", procs, float64(l1)/float64(l2))
	fmt.Printf("  Ускорение %d горутин vs 1: %.2fx\n", procs*4, float64(l1)/float64(l3))
	fmt.Printf("  Ускорение %d горутин vs 1: %.2fx\n", procs*16, float64(l1)/float64(l4))

	// =============================================================
	// Тест 3: Смешанный — мелкие + крупные одновременно
	// =============================================================
	fmt.Println("\n=== Тест 3: Смешанный (мелкие + крупные параллельно) ===")
	fmt.Println("  Горутины с мелкими аллокациями не должны мешать друг другу,")
	fmt.Println("  но крупные создают contention на mheap.")
	fmt.Println()

	var wg sync.WaitGroup
	runtime.GC()

	start := time.Now()
	// Половина горутин — мелкие аллокации
	for i := 0; i < procs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allocSmall(totalSmall / procs)
		}()
	}
	// Половина горутин — крупные аллокации
	for i := 0; i < procs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allocLarge(totalLarge / procs)
		}()
	}
	wg.Wait()
	mixed := time.Since(start)

	fmt.Printf("  Смешанный: %d горутин small + %d горутин large = %v\n",
		procs, procs, mixed)
	fmt.Printf("  Small отдельно: %v, Large отдельно: %v\n", t2, l2)
	fmt.Println()
	fmt.Println("  → Если mixed ≈ max(small, large) — аллокаторы не мешают друг другу.")
	fmt.Println("  → Если mixed > small+large — значит есть contention.")
}

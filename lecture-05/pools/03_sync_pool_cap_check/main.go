// Пример 3: sync.Pool с проверкой capacity.
//
// Улучшенный вариант: при возврате в пул проверяем cap слайса.
// Если он вырос больше допустимого максимума — выбрасываем,
// пусть GC его уберёт. Так пул не раздувается.
//
// Это стандартная практика в production-коде (net/http, prometheus и др.).
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"sync"
)

const (
	defaultBufSize = 128
	maxBufCap      = 4096 // не храним в пуле буферы > 4 КБ
)

// BufPool — обёртка над sync.Pool с контролем capacity.
type BufPool struct {
	pool sync.Pool
	max  int
}

// NewBufPool — создаёт пул с заданным лимитом capacity.
func NewBufPool(defaultSize, maxCap int) *BufPool {
	bp := &BufPool{max: maxCap}
	bp.pool.New = func() any {
		buf := make([]byte, 0, defaultSize)
		return &buf
	}
	return bp
}

// Get — берём буфер из пула.
func (bp *BufPool) Get() *[]byte {
	return bp.pool.Get().(*[]byte)
}

// Put — возвращаем буфер. Если cap > max, выбрасываем.
func (bp *BufPool) Put(bufPtr *[]byte) {
	buf := *bufPtr
	if cap(buf) > bp.max {
		// Буфер слишком вырос — не возвращаем, GC подберёт.
		return
	}
	*bufPtr = buf[:0]
	bp.pool.Put(bufPtr)
}

func main() {
	fmt.Println("=== sync.Pool с проверкой capacity ===")
	fmt.Printf("  default cap = %d, max cap = %d\n\n", defaultBufSize, maxBufCap)

	pool := NewBufPool(defaultBufSize, maxBufCap)

	// Сценарий 1: буфер не вырос — возвращается в пул.
	fmt.Println("--- Сценарий 1: нормальное использование ---")
	bp := pool.Get()
	*bp = append(*bp, make([]byte, 100)...)
	fmt.Printf("  Использовали: len=%d, cap=%d (cap <= %d → вернётся)\n",
		len(*bp), cap(*bp), maxBufCap)
	pool.Put(bp)

	bp2 := pool.Get()
	fmt.Printf("  Повторный Get: len=%d, cap=%d ← тот же буфер\n",
		len(*bp2), cap(*bp2))
	pool.Put(bp2)

	// Сценарий 2: буфер вырос — выбрасывается.
	fmt.Println("\n--- Сценарий 2: буфер вырос слишком сильно ---")
	bp3 := pool.Get()
	// Растим далеко за лимит.
	for range 50_000 {
		*bp3 = append(*bp3, 0xAB)
	}
	fmt.Printf("  Использовали: len=%d, cap=%d (cap > %d → выбросим)\n",
		len(*bp3), cap(*bp3), maxBufCap)
	pool.Put(bp3) // внутри: cap > max → не кладём

	bp4 := pool.Get()
	fmt.Printf("  Повторный Get: len=%d, cap=%d ← НОВЫЙ буфер (New())\n",
		len(*bp4), cap(*bp4))
	pool.Put(bp4)

	// Сценарий 3: массовый тест — сколько «плохих» буферов отсеялось.
	fmt.Println("\n--- Сценарий 3: массовый тест (1000 буферов) ---")

	var wg sync.WaitGroup
	var accepted, rejected int64
	var mu sync.Mutex

	for i := range 1000 {
		id := i
		wg.Go(func() {
			bp := pool.Get()

			// Чётные горутины растят буфер «в меру», нечётные — сильно.
			if id%2 == 0 {
				*bp = append(*bp, make([]byte, 200)...) // cap останется < max
			} else {
				*bp = append(*bp, make([]byte, 50_000)...) // cap >> max
			}

			grewTooMuch := cap(*bp) > maxBufCap
			pool.Put(bp)

			mu.Lock()
			if grewTooMuch {
				rejected++
			} else {
				accepted++
			}
			mu.Unlock()
		})
	}
	wg.Wait()

	fmt.Printf("  Возвращено в пул:  %d (cap <= %d)\n", accepted, maxBufCap)
	fmt.Printf("  Выброшено (GC):    %d (cap > %d)\n", rejected, maxBufCap)
	fmt.Println("\n  → Вывод: проверка cap при Put — простая и эффективная защита от утечки.")
}

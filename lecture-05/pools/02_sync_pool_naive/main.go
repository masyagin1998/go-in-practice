// Пример 2: Наивное использование sync.Pool — утечка capacity.
//
// sync.Pool позволяет переиспользовать объекты между аллокациями.
// Наивный подход: кладём слайс обратно как есть после использования.
//
// Проблема: вызывающий код может делать append и растить слайс.
// После возврата в пул, выросший слайс сохраняет свой capacity.
// Следующий Get() получит слайс с огромным capacity.
// В итоге sync.Pool тихо хранит гораздо больше памяти, чем ожидалось.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	fmt.Println("=== Наивный sync.Pool: утечка capacity ===")

	pool := sync.Pool{
		New: func() any {
			// Выделяем скромный буфер на 64 байта.
			buf := make([]byte, 0, 64)
			return &buf
		},
	}

	// Раунд 1: берём буфер, смотрим его размер.
	bufPtr := pool.Get().(*[]byte)
	buf := *bufPtr
	fmt.Printf("\n  [Раунд 1] Получили буфер: len=%d, cap=%d\n", len(buf), cap(buf))

	// Растим буфер — append аллоцирует новый backing array.
	for range 10_000 {
		buf = append(buf, 0xFF)
	}
	fmt.Printf("  [Раунд 1] После append:   len=%d, cap=%d\n", len(buf), cap(buf))

	// Наивно возвращаем — НЕ сбрасываем capacity.
	*bufPtr = buf[:0] // длину обнулили, но cap остался огромным!
	pool.Put(bufPtr)
	fmt.Printf("  [Раунд 1] Вернули:        len=%d, cap=%d\n", len(*bufPtr), cap(*bufPtr))

	// Раунд 2: берём «тот же» буфер обратно.
	bufPtr2 := pool.Get().(*[]byte)
	buf2 := *bufPtr2
	fmt.Printf("\n  [Раунд 2] Получили буфер: len=%d, cap=%d  ← cap вырос!\n",
		len(buf2), cap(buf2))

	// Масштабная демонстрация: 1000 горутин растят буферы.
	fmt.Println("\n=== Масштабный тест: 1000 горутин ===")

	bigPool := sync.Pool{
		New: func() any {
			buf := make([]byte, 0, 128)
			return &buf
		},
	}

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	var wg sync.WaitGroup
	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			bp := bigPool.Get().(*[]byte)
			b := *bp

			// Каждая горутина растит буфер до ~100 КБ.
			for range 100_000 {
				b = append(b, 0xAB)
			}

			*bp = b[:0] // длину обнулили, cap ~100 КБ остался
			bigPool.Put(bp)
		}()
	}
	wg.Wait()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	// Проверяем: что лежит в пуле?
	fmt.Println("\n  Содержимое пула (первые 10 буферов):")
	totalCap := 0
	for i := range 10 {
		obj := bigPool.Get()
		if obj == nil {
			fmt.Printf("  [%d] пул пуст\n", i)
			break
		}
		bp := obj.(*[]byte)
		fmt.Printf("  [%d] len=%-6d cap=%-8d (ожидали cap=128)\n", i, len(*bp), cap(*bp))
		totalCap += cap(*bp)
	}

	fmt.Printf("\n  Суммарный cap первых 10 буферов: %d байт (~%d КБ)\n",
		totalCap, totalCap/1024)
	fmt.Printf("  Ожидали:                         %d байт (10 × 128)\n", 10*128)
	fmt.Printf("  Перерасход:                      %.0fx\n\n",
		float64(totalCap)/float64(10*128))

	fmt.Printf("  HeapAlloc до:    %d КБ\n", before.HeapAlloc/1024)
	fmt.Printf("  HeapAlloc после: %d КБ\n", after.HeapAlloc/1024)
	fmt.Println("\n  → Вывод: наивный Put без контроля capacity превращает")
	fmt.Println("    sync.Pool в тихий «пожиратель» памяти.")
}

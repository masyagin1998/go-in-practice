// Пример 3: Бенчмарк — арена vs обычный heap.
//
// Сравниваем два сценария обработки «запроса»:
//   1. Обычные аллокации на куче + GC.
//   2. Аллокации в арене + Free().
//
// Арена выигрывает, когда:
//   - Много мелких аллокаций с коротким временем жизни.
//   - Все объекты можно освободить одновременно.
//   - GC-pressure критичен (latency-sensitive код).
//
// Арена проигрывает, когда:
//   - Объекты живут долго и по-разному.
//   - Объектов мало (overhead арены не окупается).
//   - Объекты нужно передать за пределы «запроса».
//
// Запуск:
//   GOEXPERIMENT=arenas go run main.go

package main

import (
	"arena"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

type Node struct {
	Value    float64
	Children [4]*Node
}

const (
	requestCount = 50
	nodesPerReq  = 50_000
)

// simulateRequestHeap — обработка «запроса» с обычными аллокациями.
func simulateRequestHeap() {
	nodes := make([]*Node, nodesPerReq)
	for i := range nodes {
		nodes[i] = &Node{Value: float64(i)}
		if i >= 4 {
			nodes[i].Children = [4]*Node{
				nodes[i-1], nodes[i-2], nodes[i-3], nodes[i-4],
			}
		}
	}
	// Объекты станут мусором после выхода из функции.
	runtime.KeepAlive(nodes)
}

// simulateRequestArena — обработка «запроса» с ареной.
func simulateRequestArena() {
	a := arena.NewArena()
	nodes := arena.MakeSlice[*Node](a, nodesPerReq, nodesPerReq)
	for i := range nodes {
		nodes[i] = arena.New[Node](a)
		nodes[i].Value = float64(i)
		if i >= 4 {
			nodes[i].Children = [4]*Node{
				nodes[i-1], nodes[i-2], nodes[i-3], nodes[i-4],
			}
		}
	}
	// Вся память освобождается за один вызов.
	a.Free()
}

func main() {
	// Отключаем GC для чистоты первого замера.
	debug.SetGCPercent(-1)

	// --- Heap ---
	runtime.GC()
	var mBefore runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	start := time.Now()

	for range requestCount {
		simulateRequestHeap()
	}

	heapDur := time.Since(start)
	// Включаем GC и замеряем время сборки мусора.
	debug.SetGCPercent(100)
	gcStart := time.Now()
	runtime.GC()
	gcDur := time.Since(gcStart)

	var mAfterGC runtime.MemStats
	runtime.ReadMemStats(&mAfterGC)

	fmt.Printf("=== Heap + GC ===\n")
	fmt.Printf("Обработка %d запросов: %v\n", requestCount, heapDur)
	fmt.Printf("Финальный GC:          %v\n", gcDur)
	fmt.Printf("Всего:                 %v\n", heapDur+gcDur)
	fmt.Printf("GC-циклов (NumGC):     %d\n", mAfterGC.NumGC-mBefore.NumGC)
	fmt.Printf("PauseTotalNs:          %v\n\n", time.Duration(mAfterGC.PauseTotalNs-mBefore.PauseTotalNs))

	// --- Arena ---
	runtime.GC()
	debug.FreeOSMemory()
	runtime.ReadMemStats(&mBefore)
	start = time.Now()

	for range requestCount {
		simulateRequestArena()
	}

	arenaDur := time.Since(start)

	var mAfterArena runtime.MemStats
	runtime.ReadMemStats(&mAfterArena)

	fmt.Printf("=== Arena ===\n")
	fmt.Printf("Обработка %d запросов: %v\n", requestCount, arenaDur)
	fmt.Printf("GC-циклов (NumGC):     %d\n", mAfterArena.NumGC-mBefore.NumGC)
	fmt.Printf("PauseTotalNs:          %v\n\n", time.Duration(mAfterArena.PauseTotalNs-mBefore.PauseTotalNs))

	// --- Итоги ---
	fmt.Printf("=== Сравнение ===\n")
	fmt.Printf("Heap + GC: %v\n", heapDur+gcDur)
	fmt.Printf("Arena:     %v\n", arenaDur)
	if arenaDur < heapDur+gcDur {
		fmt.Printf("Арена быстрее в %.1fx\n", float64(heapDur+gcDur)/float64(arenaDur))
	} else {
		fmt.Printf("Heap быстрее в %.1fx\n", float64(arenaDur)/float64(heapDur+gcDur))
	}
}

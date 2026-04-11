// Пример 3: Стек vs куча — разница в производительности.
//
// Аллокация на стеке — это просто сдвиг указателя (SP), практически бесплатно.
// Аллокация в куче — вызов runtime.mallocgc, плюс нагрузка на GC при сборке.
// Этот бенчмарк показывает разницу наглядно.
//
// Запуск:
//   go run main.go
//
// Или как бенчмарк:
//   go test -bench=. -benchmem

package main

import (
	"fmt"
	"runtime"
	"time"
)

type Data struct {
	Values [64]int
}

// allocOnStack — структура не убегает, живёт на стеке.
// Стоимость: ~0 нс (сдвиг SP).
//
//go:noinline
func allocOnStack() int {
	d := Data{}
	d.Values[0] = 42
	return d.Values[0]
}

// allocOnHeap — структура убегает через указатель, живёт в куче.
// Стоимость: вызов mallocgc + давление на GC.
//
//go:noinline
func allocOnHeap() *Data {
	d := &Data{}
	d.Values[0] = 42
	return d
}

func main() {
	const N = 10_000_000

	// Прогрев.
	for range 1000 {
		allocOnStack()
		allocOnHeap()
	}
	runtime.GC()

	// Бенчмарк: стек.
	var statsBeforeStack runtime.MemStats
	runtime.ReadMemStats(&statsBeforeStack)

	start := time.Now()
	for range N {
		allocOnStack()
	}
	stackDur := time.Since(start)

	var statsAfterStack runtime.MemStats
	runtime.ReadMemStats(&statsAfterStack)

	runtime.GC()

	// Бенчмарк: куча.
	var statsBeforeHeap runtime.MemStats
	runtime.ReadMemStats(&statsBeforeHeap)

	start = time.Now()
	for range N {
		allocOnHeap()
	}
	heapDur := time.Since(start)

	var statsAfterHeap runtime.MemStats
	runtime.ReadMemStats(&statsAfterHeap)

	fmt.Printf("=== %d итераций ===\n\n", N)
	fmt.Printf("Стек:  %v, аллокаций в куче: %d, память: %d КБ\n",
		stackDur,
		statsAfterStack.Mallocs-statsBeforeStack.Mallocs,
		(statsAfterStack.TotalAlloc-statsBeforeStack.TotalAlloc)/1024)

	fmt.Printf("Куча:  %v, аллокаций в куче: %d, память: %d КБ\n",
		heapDur,
		statsAfterHeap.Mallocs-statsBeforeHeap.Mallocs,
		(statsAfterHeap.TotalAlloc-statsBeforeHeap.TotalAlloc)/1024)

	fmt.Printf("\nКуча медленнее в %.1fx\n", float64(heapDur)/float64(stackDur))
	fmt.Printf("Циклов GC за всё время: %d\n", statsAfterHeap.NumGC)
}

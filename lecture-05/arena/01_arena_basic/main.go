// Пример 1: Базовое использование arena в Go.
//
// Пакет arena — экспериментальная фича (GOEXPERIMENT=arenas).
// Позволяет выделять память для группы объектов и освобождать её
// целиком за один вызов Free(), не дожидаясь сборщика мусора.
//
// Идея: арена выделяет большой блок памяти, из которого «нарезаются»
// объекты. Когда все объекты больше не нужны — Free() освобождает
// весь блок разом. Это аналог region-based memory management.
//
// Демонстрируем:
//   1. Создание арены (arena.NewArena).
//   2. Аллокацию объектов через arena.New[T].
//   3. Освобождение арены через Free().
//   4. Сравнение потребления памяти: арена vs обычный heap.
//
// Запуск:
//   GOEXPERIMENT=arenas go run main.go

package main

import (
	"arena"
	"fmt"
	"runtime"
	"runtime/debug"
)

// Vertex — точка в 3D-пространстве.
type Vertex struct {
	X, Y, Z float64
	Normal   [3]float64
	Color    [4]byte
}

func memUsage() uint64 {
	runtime.GC()
	debug.FreeOSMemory()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapInuse
}

func main() {
	const N = 100_000

	// --- Обычные аллокации на куче ---
	before := memUsage()
	heapVertices := make([]*Vertex, N)
	for i := range heapVertices {
		heapVertices[i] = &Vertex{
			X: float64(i), Y: float64(i), Z: float64(i),
		}
	}
	var afterHeap runtime.MemStats
	runtime.ReadMemStats(&afterHeap)
	fmt.Printf("=== Обычный heap ===\n")
	fmt.Printf("Аллоцировано %d объектов Vertex\n", N)
	fmt.Printf("HeapInuse: %d КБ\n\n", (afterHeap.HeapInuse-before)/1024)

	// Освобождаем ссылки, чтобы GC мог собрать.
	heapVertices = nil
	runtime.GC()
	debug.FreeOSMemory()

	// --- Аллокации через арену ---
	before = memUsage()
	a := arena.NewArena()

	arenaVertices := make([]*Vertex, N)
	for i := range arenaVertices {
		// arena.New[T] — обобщённая функция, создаёт объект типа T
		// в памяти арены. Возвращает *T.
		v := arena.New[Vertex](a)
		v.X = float64(i)
		v.Y = float64(i)
		v.Z = float64(i)
		arenaVertices[i] = v
	}
	var afterArena runtime.MemStats
	runtime.ReadMemStats(&afterArena)
	fmt.Printf("=== Arena ===\n")
	fmt.Printf("Аллоцировано %d объектов Vertex в арене\n", N)
	fmt.Printf("HeapInuse: %d КБ\n", (afterArena.HeapInuse-before)/1024)
	fmt.Printf("Vertex[0] = {X: %.0f, Y: %.0f, Z: %.0f}\n", arenaVertices[0].X, arenaVertices[0].Y, arenaVertices[0].Z)
	fmt.Printf("Vertex[%d] = {X: %.0f, Y: %.0f, Z: %.0f}\n\n", N-1, arenaVertices[N-1].X, arenaVertices[N-1].Y, arenaVertices[N-1].Z)

	// Free() освобождает всю память арены разом.
	// После вызова все указатели, полученные из арены, становятся невалидными.
	// Обращение к ним вызовет fault (use-after-free защита).
	arenaVertices = nil
	a.Free()

	afterFree := memUsage()
	fmt.Printf("После arena.Free():\n")
	fmt.Printf("HeapInuse: %d КБ\n", afterFree/1024)
	fmt.Println("Память арены помечена как свободная без ожидания GC-цикла.")
	fmt.Println("Runtime может не сразу вернуть страницы ОС (аналогично madvise(MADV_FREE)).")
}

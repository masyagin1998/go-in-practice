// Пример 4: Арена и вложенные структуры — подводный камень.
//
// arena.New[T] выделяет в арене только сам объект T.
// Если T содержит указатели, слайсы или map — они НЕ попадают в арену
// автоматически. Вложенные ссылочные поля аллоцируются на обычной куче.
//
// Это значит:
//   - arena.New[Node] → сам Node в арене.
//   - node.Children = append(...) → backing array слайса на куче.
//   - node.Name = fmt.Sprintf(...) → строка на куче.
//   - node.Meta = &Meta{...} → Meta на куче, НЕ в арене.
//
// Чтобы всё было в арене, нужно явно выделять каждый вложенный объект
// через arena.New / arena.MakeSlice.
//
// Демонстрируем:
//   1. «Наивный» подход — вложенные объекты утекают на кучу.
//   2. «Правильный» подход — явное выделение всего в арене.
//   3. Сравнение HeapAlloc: сколько мусора на куче в каждом случае.
//
// Запуск:
//   GOEXPERIMENT=arenas go run main.go

package main

import (
	"arena"
	"fmt"
	"runtime"
)

type Meta struct {
	CreatedBy string
	Version   int
}

type Node struct {
	ID       int
	Name     string
	Meta     *Meta
	Children []*Node
}

func heapAlloc() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.TotalAlloc
}

func main() {
	const N = 10_000

	// =============================================
	// 1. Наивный подход: только верхний объект в арене
	// =============================================
	runtime.GC()
	before := heapAlloc()

	a1 := arena.NewArena()
	root := arena.New[Node](a1)
	root.ID = 0
	root.Name = "root" // строковый литерал — не аллокация

	for i := 1; i <= N; i++ {
		child := arena.New[Node](a1) // сам Node — в арене ✓
		child.ID = i
		child.Name = fmt.Sprintf("node_%d", i) // строка → КУЧА ✗
		child.Meta = &Meta{                     // Meta → КУЧА ✗
			CreatedBy: fmt.Sprintf("gen_%d", i), // строка → КУЧА ✗
			Version:   i,
		}
		// append → backing array Children → КУЧА ✗
		root.Children = append(root.Children, child)
	}

	after := heapAlloc()
	a1.Free()

	naiveHeap := after - before
	fmt.Printf("=== Наивный подход ===\n")
	fmt.Printf("Node в арене, но вложенные объекты на куче.\n")
	fmt.Printf("Heap-аллокаций: %d КБ\n\n", naiveHeap/1024)

	// =============================================
	// 2. Правильный подход: всё явно в арене
	// =============================================
	runtime.GC()
	before = heapAlloc()

	a2 := arena.NewArena()
	root2 := arena.New[Node](a2)
	root2.ID = 0
	root2.Name = "root"
	// Выделяем слайс children в арене заранее.
	root2.Children = arena.MakeSlice[*Node](a2, 0, N) // backing array → АРЕНА ✓

	for i := 1; i <= N; i++ {
		child := arena.New[Node](a2)  // Node → АРЕНА ✓
		child.ID = i
		child.Name = fmt.Sprintf("node_%d", i) // строка → КУЧА ✗ (строки нельзя выделить в арене)

		meta := arena.New[Meta](a2)             // Meta → АРЕНА ✓
		meta.CreatedBy = fmt.Sprintf("gen_%d", i) // строка → КУЧА ✗
		meta.Version = i
		child.Meta = meta

		root2.Children = append(root2.Children, child) // append в арена-слайс → АРЕНА ✓
	}

	after = heapAlloc()
	a2.Free()

	properHeap := after - before
	fmt.Printf("=== Правильный подход ===\n")
	fmt.Printf("Node, Meta и backing array Children — всё в арене.\n")
	fmt.Printf("На куче остались только строки (fmt.Sprintf).\n")
	fmt.Printf("Heap-аллокаций: %d КБ\n\n", properHeap/1024)

	// =============================================
	// 3. Итоги
	// =============================================
	fmt.Printf("=== Сравнение ===\n")
	fmt.Printf("Наивный:    %d КБ на куче\n", naiveHeap/1024)
	fmt.Printf("Правильный: %d КБ на куче\n", properHeap/1024)
	saved := float64(naiveHeap-properHeap) / float64(naiveHeap) * 100
	fmt.Printf("Экономия:   %.1f%% heap-аллокаций\n\n", saved)

	fmt.Println("Вывод: arena.New[T] выделяет только сам T.")
	fmt.Println("Указатели, слайсы и map внутри T — обычная куча.")
	fmt.Println("Для полного контроля — arena.New / arena.MakeSlice для каждого поля.")
	fmt.Println("Строки — исключение: их нельзя выделить в арене напрямую.")
}

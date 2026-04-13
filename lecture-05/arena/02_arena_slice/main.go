// Пример 2: Создание слайсов в арене через arena.MakeSlice.
//
// arena.MakeSlice[T](a, len, cap) — аналог make([]T, len, cap),
// но backing array размещается в памяти арены.
//
// Типичный use-case: обработка запроса, где создаётся много
// временных слайсов. Вместо того чтобы GC собирал их после
// каждого запроса, можно выделить всё в арене и освободить разом.
//
// Также демонстрируем arena.Clone — копирование значения
// из арены в обычную кучу (чтобы оно пережило Free).
//
// Запуск:
//   GOEXPERIMENT=arenas go run main.go

package main

import (
	"arena"
	"fmt"
	"strings"
)

// Record — запись из «базы данных».
type Record struct {
	ID   int
	Name string
	Tags []string
}

func main() {
	// Имитируем обработку запроса: создаём много объектов,
	// которые нужны только на время обработки.
	a := arena.NewArena()

	// --- MakeSlice: слайс в арене ---
	// Backing array слайса размещён в арене, а не на куче.
	records := arena.MakeSlice[Record](a, 0, 1000)

	for i := range 1000 {
		// Каждый Record добавляется в слайс, который живёт в арене.
		// Но строки (Name, Tags) — это обычные Go-строки на куче,
		// арена управляет только самими структурами Record.
		records = append(records, Record{
			ID:   i,
			Name: fmt.Sprintf("user_%d", i),
			Tags: strings.Split("go,arena,memory", ","),
		})
	}

	fmt.Printf("Создано %d записей в арене\n", len(records))
	fmt.Printf("Первая: %+v\n", records[0])
	fmt.Printf("Последняя: %+v\n\n", records[len(records)-1])

	// --- arena.Clone: спасаем значение из арены ---
	// Если нужно сохранить объект после Free(), используем Clone.
	// Clone работает с указателями, слайсами и строками.
	// Для структуры — берём указатель на элемент слайса.
	savedPtr := arena.Clone(&records[0])
	fmt.Printf("Клонировано в кучу: %+v\n\n", *savedPtr)

	// --- Множественные слайсы в одной арене ---
	ids := arena.MakeSlice[int](a, 0, 1000)
	names := arena.MakeSlice[string](a, 0, 1000)
	for _, r := range records {
		ids = append(ids, r.ID)
		names = append(names, r.Name)
	}
	fmt.Printf("Слайс ID:    len=%d, cap=%d\n", len(ids), cap(ids))
	fmt.Printf("Слайс Names: len=%d, cap=%d\n\n", len(names), cap(names))

	// Free() — один вызов освобождает ВСЮ память арены:
	// и records, и ids, и names — всё за O(1).
	a.Free()
	fmt.Println("Арена освобождена. Все слайсы невалидны.")
	fmt.Printf("Но клонированная запись жива: %+v\n", *savedPtr)
}

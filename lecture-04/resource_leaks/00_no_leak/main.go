// Пример 0: Программа БЕЗ утечек.
// Выделяем много памяти в цикле, но GC всё подчищает.
//
// Примечание: поддержка valgrind в Go экспериментальная (go#73602).
// Valgrind может показать ошибки "Invalid read" из runtime (GC mark phase,
// планировщик горутин) — это ложные срабатывания из-за неполной аннотации
// рантайма (трекается в go#73801), а не утечки нашего кода.
//
// Сборка:
//   go build -tags valgrind -o 00_no_leak .
//
// Запуск:
//   ./00_no_leak
//
// Запуск с valgrind:
//   valgrind --tool=memcheck ./00_no_leak

package main

import (
	"fmt"
	"runtime"
)

func main() {
	var stats runtime.MemStats

	// Выделяем 10 000 слайсов по 1 КБ — суммарно ~10 МБ аллокаций.
	// Каждый слайс становится мусором на следующей итерации,
	// и GC его подбирает.
	// (Под valgrind программа работает в ~50 раз медленнее,
	// поэтому ограничиваемся 10 000 итераций.)
	for i := range 10_000 {
		buf := make([]byte, 1024)
		buf[0] = byte(i) // Используем, чтобы компилятор не выкинул аллокацию.
		_ = buf
	}

	// Принудительно запускаем GC, чтобы подчистить остатки.
	runtime.GC()

	runtime.ReadMemStats(&stats)
	fmt.Printf("Всего аллоцировано: %d МБ\n", stats.TotalAlloc/1024/1024)
	fmt.Printf("Сейчас в куче:      %d КБ\n", stats.HeapInuse/1024)
	fmt.Printf("Циклов GC:          %d\n", stats.NumGC)
	fmt.Println("Утечек нет — GC всё подчистил.")
}

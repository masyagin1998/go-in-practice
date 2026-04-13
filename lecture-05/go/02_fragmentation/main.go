// Пример 2: Фрагментация памяти в Go.
//
// Демонстрируем два вида фрагментации:
//   1. Внутренняя (internal) — потери на округление до size-класса.
//   2. Внешняя (external) — span'ы частично заняты, свободные слоты не используются.
//
// Сценарий:
//   - Выделяем N объектов
//   - Освобождаем каждый второй (имитируем «швейцарский сыр»)
//   - Смотрим HeapInuse vs HeapAlloc — разница = фрагментация
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Структура с «неудобным» размером — 33 байта.
// Попадёт в size-класс 48 (потери: 15 байт = 31%).
type awkwardObj struct {
	data [33]byte
}

// Структура ровно 64 байта — идеально ложится в size-класс 64.
type niceObj struct {
	data [64]byte
}

// memStatsNoGC — читаем статистику без вызова GC (чтобы не собрать живые объекты).
func memStatsNoGC() runtime.MemStats {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms
}

// memStatsWithGC — GC + FreeOSMemory + статистика.
func memStatsWithGC() runtime.MemStats {
	runtime.GC()
	debug.FreeOSMemory()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms
}

func printStats(label string, ms runtime.MemStats) {
	frag := float64(0)
	if ms.HeapInuse > 0 {
		frag = float64(ms.HeapInuse-ms.HeapAlloc) / float64(ms.HeapInuse) * 100
	}
	fmt.Printf("  %-35s HeapAlloc=%-8d HeapInuse=%-8d Objects=%-6d Frag=%.1f%%\n",
		label, ms.HeapAlloc, ms.HeapInuse, ms.HeapObjects, frag)
}

func main() {
	const N = 10_000

	// ============================================================
	// Сценарий 1: Внутренняя фрагментация (округление до size-класса)
	// ============================================================
	fmt.Println("=== Сценарий 1: Внутренняя фрагментация ===")
	fmt.Printf("  awkwardObj: запрошено 33 байта → size-класс 48 (потери 31%%)\n")
	fmt.Printf("  niceObj:    запрошено 64 байта → size-класс 64 (потери 0%%)\n\n")

	runtime.GC()
	base := memStatsNoGC()

	// Выделяем awkwardObj (33 байта → 48 в аллокаторе)
	awkward := make([]*awkwardObj, N)
	for i := range awkward {
		awkward[i] = new(awkwardObj)
		awkward[i].data[0] = byte(i)
	}
	after1 := memStatsNoGC()
	printStats("[после 10K awkwardObj (33→48)]:", after1)
	fmt.Printf("  Данных по факту:   %d КБ (10K × 33)\n", N*33/1024)
	fmt.Printf("  HeapAlloc прирост: %d КБ (10K × 48 = %d КБ ожидаемо)\n",
		(after1.HeapAlloc-base.HeapAlloc)/1024, N*48/1024)
	runtime.KeepAlive(awkward)
	awkward = nil

	runtime.GC()
	clean1 := memStatsNoGC()

	// Выделяем niceObj (64 байта → 64 в аллокаторе)
	nice := make([]*niceObj, N)
	for i := range nice {
		nice[i] = new(niceObj)
		nice[i].data[0] = byte(i)
	}
	after2 := memStatsNoGC()
	printStats("[после 10K niceObj (64→64)]:", after2)
	fmt.Printf("  Данных по факту:   %d КБ (10K × 64)\n", N*64/1024)
	fmt.Printf("  HeapAlloc прирост: %d КБ (10K × 64 = %d КБ ожидаемо)\n",
		(after2.HeapAlloc-clean1.HeapAlloc)/1024, N*64/1024)
	runtime.KeepAlive(nice)
	nice = nil

	// ============================================================
	// Сценарий 2: Внешняя фрагментация («швейцарский сыр»)
	// ============================================================
	fmt.Println("\n=== Сценарий 2: Внешняя фрагментация (\"швейцарский сыр\") ===")

	_ = memStatsWithGC()

	// Выделяем N объектов
	objs := make([]*[128]byte, N)
	for i := range objs {
		objs[i] = new([128]byte)
		objs[i][0] = byte(i)
	}
	full := memStatsNoGC()
	printStats("[все 10K объектов живы]:", full)

	// Освобождаем каждый второй — создаём дырки
	for i := 0; i < N; i += 2 {
		objs[i] = nil
	}
	holed := memStatsWithGC()
	printStats("[каждый второй = nil]:", holed)
	fmt.Println()
	fmt.Println("  → HeapInuse почти не уменьшился, хотя половина объектов мертва.")
	fmt.Println("    Span не может быть освобождён, пока в нём жив хотя бы 1 объект.")
	fmt.Println("    Это и есть внешняя фрагментация.")

	// Освобождаем оставшиеся
	for i := 1; i < N; i += 2 {
		objs[i] = nil
	}
	empty := memStatsWithGC()
	printStats("[все объекты = nil]:", empty)
	fmt.Println()
	fmt.Println("  → Теперь span'ы полностью пусты и могут быть возвращены.")

	// ============================================================
	// Сценарий 3: Смешанные размеры
	// ============================================================
	fmt.Println("\n=== Сценарий 3: Смешанные размеры ===")
	fmt.Println("  Чередуем маленькие (16 байт) и средние (512 байт) объекты,")
	fmt.Println("  потом убиваем маленькие. У Go разные size-классы не мешают")
	fmt.Println("  друг другу — в отличие от C-аллокаторов с единым free list.")

	_ = memStatsWithGC()

	type small struct{ x [16]byte }
	type medium struct{ x [512]byte }

	smalls := make([]*small, N)
	mediums := make([]*medium, N)

	for i := 0; i < N; i++ {
		smalls[i] = new(small)
		mediums[i] = new(medium)
	}
	mixed := memStatsNoGC()
	printStats("[10K small + 10K medium]:", mixed)

	// Убиваем все маленькие
	for i := range smalls {
		smalls[i] = nil
	}
	onlyMedium := memStatsWithGC()
	printStats("[убили все small]:", onlyMedium)
	runtime.KeepAlive(mediums)

	fmt.Printf("\n  Реально нужно medium: %d КБ (10K × 512)\n", N*512/1024)
	fmt.Printf("  HeapAlloc:            %d КБ\n", onlyMedium.HeapAlloc/1024)
	fmt.Printf("  HeapInuse:            %d КБ\n", onlyMedium.HeapInuse/1024)
	fmt.Println()
	fmt.Println("  → Small и medium живут в разных span'ах (разные size-классы).")
	fmt.Println("    Освобождение small'ов полностью освобождает их span'ы.")
	fmt.Println("    Это преимущество size-class аллокатора над единым free list.")
}

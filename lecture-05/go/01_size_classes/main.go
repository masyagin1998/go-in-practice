// Пример 1: Size-классы аллокатора Go.
//
// Go использует 67 size-классов для объектов <= 32 КБ.
// Каждый запрос округляется вверх до ближайшего класса.
// Это даёт внутреннюю фрагментацию (internal fragmentation) — разницу
// между запрошенным и реально выделенным размером.
//
// Таблица size-классов живёт в runtime/sizeclasses.go.
// Максимальные потери на округление — 12.5% (по дизайну).
//
// Здесь мы выделяем объекты разных размеров и через unsafe определяем,
// в какой size-класс они попали, измеряя «дырки» между аллокациями.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

// sizeClassTable — таблица size-классов Go (из runtime/sizeclasses.go).
// class → (размер объекта, размер span в байтах, объектов в span).
var sizeClassTable = []struct {
	Class      int
	ObjSize    int
	SpanSize   int
	ObjPerSpan int
}{
	{1, 8, 8192, 1024},
	{2, 16, 8192, 512},
	{3, 24, 8192, 341},
	{4, 32, 8192, 256},
	{5, 48, 8192, 170},
	{6, 64, 8192, 128},
	{7, 80, 8192, 102},
	{8, 96, 8192, 85},
	{9, 112, 8192, 73},
	{10, 128, 8192, 64},
	{11, 144, 8192, 56},
	{12, 160, 8192, 51},
	{13, 176, 8192, 46},
	{14, 192, 8192, 42},
	{15, 208, 8192, 39},
	{16, 224, 8192, 36},
	{17, 240, 8192, 34},
	{18, 256, 8192, 32},
	{19, 288, 8192, 28},
	{20, 320, 8192, 25},
	{21, 352, 8192, 23},
	{22, 384, 8192, 21},
	{23, 416, 8192, 19},
	{24, 448, 8192, 18},
	{25, 480, 8192, 17},
	{26, 512, 8192, 16},
	{27, 576, 8192, 14},
	{28, 640, 8192, 12},
	{29, 704, 8192, 11},
	{30, 768, 8192, 10},
	{31, 896, 8192, 9},
	{32, 1024, 8192, 8},
	{33, 1152, 8192, 7},
	{34, 1280, 8192, 6},
	{35, 1408, 16384, 11},
	{36, 1536, 8192, 5},
	{37, 1792, 16384, 9},
	{38, 2048, 8192, 4},
	{39, 2304, 16384, 7},
	{40, 2688, 8192, 3},
	{41, 3072, 24576, 8},
	{42, 3200, 16384, 5},
	{43, 3456, 24576, 7},
	{44, 4096, 8192, 2},
	{45, 4864, 24576, 5},
	{46, 5376, 16384, 3},
	{47, 6144, 24576, 4},
	{48, 6528, 32768, 5},
	{49, 6784, 40960, 6},
	{50, 6912, 49152, 7},
	{51, 8192, 8192, 1},
	{52, 9472, 57344, 6},
	{53, 9728, 49152, 5},
	{54, 10240, 40960, 4},
	{55, 10880, 32768, 3},
	{56, 12288, 24576, 2},
	{57, 13568, 40960, 3},
	{58, 14336, 57344, 4},
	{59, 16384, 16384, 1},
	{60, 18432, 73728, 4},
	{61, 19072, 57344, 3},
	{62, 20480, 40960, 2},
	{63, 21760, 65536, 3},
	{64, 24576, 24576, 1},
	{65, 27264, 81920, 3},
	{66, 28672, 57344, 2},
	{67, 32768, 32768, 1},
}

// findSizeClass — находит size-класс для заданного размера.
func findSizeClass(size int) (class, objSize int) {
	for _, sc := range sizeClassTable {
		if sc.ObjSize >= size {
			return sc.Class, sc.ObjSize
		}
	}
	return 0, size // > 32KB — большой объект
}

func main() {
	// Часть 1: Печатаем таблицу size-классов
	fmt.Println("=== Таблица size-классов Go (выборочно) ===")
	fmt.Printf("%-6s %-12s %-12s %-12s %-10s\n",
		"Class", "ObjSize", "SpanSize", "Obj/Span", "Waste%")
	fmt.Println("-----------------------------------------------------------")

	for _, sc := range sizeClassTable {
		tailWaste := sc.SpanSize - sc.ObjSize*sc.ObjPerSpan
		wastePercent := float64(tailWaste) / float64(sc.SpanSize) * 100
		if sc.Class <= 10 || sc.Class == 32 || sc.Class == 51 || sc.Class == 67 {
			fmt.Printf("%-6d %-12d %-12d %-12d %.2f%%\n",
				sc.Class, sc.ObjSize, sc.SpanSize, sc.ObjPerSpan, wastePercent)
		}
	}

	// Часть 2: Демонстрация округления размеров
	fmt.Println("\n=== Округление размеров до size-класса ===")
	fmt.Printf("%-15s %-10s %-10s %-10s\n",
		"Запрошено", "Класс", "Реально", "Потери")
	fmt.Println("-----------------------------------------------")

	testSizes := []int{1, 7, 9, 17, 25, 33, 49, 100, 200, 500, 1000, 4000, 8000, 16000, 32000}
	for _, sz := range testSizes {
		class, objSize := findSizeClass(sz)
		waste := objSize - sz
		fmt.Printf("%-15d %-10d %-10d %-10d (%.1f%%)\n",
			sz, class, objSize, waste, float64(waste)/float64(objSize)*100)
	}

	// Часть 3: Проверяем через unsafe — реальные адреса аллокаций
	fmt.Println("\n=== Адреса аллокаций одного size-класса (16 байт) ===")
	ptrs := make([]*[16]byte, 8)
	for i := range ptrs {
		ptrs[i] = new([16]byte)
	}
	for i, p := range ptrs {
		addr := uintptr(unsafe.Pointer(p))
		if i > 0 {
			prev := uintptr(unsafe.Pointer(ptrs[i-1]))
			fmt.Printf("  [%d] addr=%#x  delta=%d байт\n", i, addr, addr-prev)
		} else {
			fmt.Printf("  [%d] addr=%#x\n", i, addr)
		}
	}
	fmt.Println("  (шаг 16 = size-класс 2, 512 объектов в span 8КБ)")

	// Часть 4: Общая статистика
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Println("\n=== MemStats ===")
	fmt.Printf("  HeapAlloc:   %d КБ (живые объекты)\n", ms.HeapAlloc/1024)
	fmt.Printf("  HeapInuse:   %d КБ (span'ы с объектами)\n", ms.HeapInuse/1024)
	fmt.Printf("  HeapIdle:    %d КБ (пустые span'ы)\n", ms.HeapIdle/1024)
	fmt.Printf("  HeapSys:     %d КБ (всего от ОС)\n", ms.HeapSys/1024)
	fmt.Printf("  Mallocs:     %d\n", ms.Mallocs)
	fmt.Printf("  Frees:       %d\n", ms.Frees)
	frag := float64(ms.HeapInuse-ms.HeapAlloc) / float64(ms.HeapInuse) * 100
	fmt.Printf("  Фрагментация (InUse-Alloc)/InUse: %.1f%%\n", frag)
}

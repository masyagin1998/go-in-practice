// Package torntest содержит общий код для демонстрации torn read.
//
// WriterThread попеременно записывает два значения в *int64.
// ReaderThread читает и проверяет, не получилось ли «смешанное» значение.
package torntest

import (
	"fmt"
	"runtime"
	"unsafe"
)

const (
	ValA = int64(0x7AAAAAAABBBBBBBB)
	ValB = int64(0x7CCCCCCCDDDDDDDD)
)

// AlignTo64 находит первый адрес >= base, выровненный по 64 (кэш-линия).
//
// Пример: base = 0x1037
//
//	base + 63       = 0x1076           // сдвигаем вперёд на максимум
//	0x1076 &^ 63    = 0x1076 & 0x...C0 // обнуляем младшие 6 бит
//	                = 0x1040           // ближайший адрес, кратный 64
func AlignTo64(base uintptr) uintptr {
	return (base + 63) &^ 63
}

// PtrAtOffset возвращает *int64 по заданному смещению от base.
func PtrAtOffset(base uintptr, offset uintptr) *int64 {
	return (*int64)(unsafe.Pointer(base + offset))
}

// store64 записывает значение по указателю.
// go:noinline не даёт компилятору заинлайнить и выкинуть «мёртвый» store.
// Без этого Go оптимизирует `*ptr = ValA; *ptr = ValB` → `*ptr = ValB`.
// В C аналог — volatile.
//
//go:noinline
func store64(ptr *int64, val int64) { *ptr = val }

//go:noinline
func load64(ptr *int64) int64 { return *ptr }

func WriterThread(ptr *int64) {
	runtime.LockOSThread()
	for {
		store64(ptr, ValA)
		store64(ptr, ValB)
	}
}

// Run запускает двух писателей и читателя, печатает результат.
// Два писателя на разных OS-тредах максимизируют давление на кэш-линию.
func Run(ptr *int64) {
	runtime.GOMAXPROCS(4)
	*ptr = ValA

	go WriterThread(ptr)
	go WriterThread(ptr)

	// Читатель — тоже на своём OS-треде.
	runtime.LockOSThread()

	tornCount := 0
	iterations := 0

	for {
		v := load64(ptr)
		iterations++
		if v != ValA && v != ValB && v != 0 {
			tornCount++
			fmt.Printf("Torn read #%d: 0x%016X (итерация %d)\n",
				tornCount, uint64(v), iterations)
			if tornCount >= 20 {
				fmt.Printf("\nИтого: %d torn reads из %d чтений.\n", tornCount, iterations)
				return
			}
		}
		if iterations%10_000_000 == 0 {
			fmt.Printf("... %d итераций, torn reads: %d\n", iterations, tornCount)
		}
	}
}

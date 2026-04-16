package main

// Передача массивов/слайсов между Go и C.
//
// Go-слайс нельзя передать напрямую — нужен указатель на первый элемент.
// Начиная с Go 1.17, unsafe.Slice позволяет удобно превратить
// C-указатель обратно в Go-слайс без ручной арифметики указателей.

/*
#include "arrays.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	// === Go → C: передаём Go-слайс как указатель + длина ===
	data := []C.int{10, 20, 30, 40, 50}
	sum := C.sum_ints(
		&data[0],          // указатель на первый элемент
		C.size_t(len(data)), // длина
	)
	fmt.Printf("Сумма [10,20,30,40,50] = %d\n", sum)

	// === C заполняет Go-слайс ===
	// Аллоцируем слайс в Go, передаём указатель в C для заполнения.
	buf := make([]C.int, 6)
	C.fill_squares(&buf[0], C.size_t(len(buf)))
	fmt.Printf("Квадраты (C заполнил Go-слайс): %v\n", buf)

	// === C-массив → Go-слайс через unsafe.Slice (Go 1.17+) ===
	// Выделяем массив в C-heap.
	n := 5
	cArr := (*C.int)(C.malloc(C.size_t(n) * C.size_t(unsafe.Sizeof(C.int(0)))))
	defer C.free(unsafe.Pointer(cArr))

	C.fill_squares(cArr, C.size_t(n))

	// unsafe.Slice создаёт Go-слайс поверх C-памяти.
	// Это удобная альтернатива reflect.SliceHeader или ручному кастингу.
	goSlice := unsafe.Slice(cArr, n)
	fmt.Printf("C-массив как Go-слайс (unsafe.Slice): %v\n", goSlice)

	// Можно итерировать как обычный слайс.
	for i, v := range goSlice {
		fmt.Printf("  [%d] = %d\n", i, v)
	}
}

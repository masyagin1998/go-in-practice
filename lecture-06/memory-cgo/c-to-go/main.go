package main

// Получение данных из C в Go:
// - простые значения (int, double) — копируются по стеку;
// - структуры — возвращаются по значению (копия на стеке Go);
// - C-массивы — приводим к Go-слайсу через unsafe.Slice.

/*
#include "producer.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	// === Простые значения ===
	// C возвращает int/double — Go получает C.int/C.double (копия).
	answer := C.get_answer()
	pi := C.get_pi()
	fmt.Printf("Ответ: %d (тип %T)\n", answer, answer)
	fmt.Printf("Пи:    %f (тип %T)\n", pi, pi)

	// === Структура по значению ===
	// C возвращает struct Color — Go получает копию C.Color на стеке.
	// Никакой аллокации в heap, никаких free().
	color := C.make_color(255, 128, 0, 255)
	fmt.Printf("\n[Go] Color: r=%d, g=%d, b=%d, a=%d\n", color.r, color.g, color.b, color.a)
	C.print_color(color)

	// Можно использовать C-структуру в Go как обычное значение.
	color2 := color
	color2.r = 0
	color2.b = 255
	fmt.Printf("[Go] Копия с изменениями: r=%d, g=%d, b=%d, a=%d\n",
		color2.r, color2.g, color2.b, color2.a)

	// === C-массив → Go-слайс ===
	// C аллоцирует массив в C-heap (malloc).
	// Go получает структуру {указатель, длина} и строит слайс через unsafe.Slice.
	fmt.Println()
	arr := C.make_fibonacci(10)
	defer C.free(unsafe.Pointer(arr.data)) // Не забываем освободить C-память!

	// unsafe.Slice (Go 1.17+) создаёт Go-слайс поверх C-памяти.
	goSlice := unsafe.Slice(arr.data, arr.len)
	fmt.Printf("Фибоначчи (C-массив как Go-слайс): %v\n", goSlice)

	// Слайс ссылается на C-память — можно итерировать, но
	// нельзя использовать после C.free().
	var sum int
	for _, v := range goSlice {
		sum += int(v)
	}
	fmt.Printf("Сумма: %d\n", sum)
}

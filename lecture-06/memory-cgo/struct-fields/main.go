package main

// Передача полей структур (не указателей) в C-функции.
// Вложенные структуры без указателей — полностью value-type,
// CGo разрешает передавать их и по значению, и по указателю.

/*
#include "rect.h"
*/
import "C"

import "fmt"

func main() {
	r := C.Rect{
		origin: C.Vec2{x: 1, y: 2},
		size:   C.Vec2{x: 10, y: 5},
	}

	// Передаём структуру целиком по значению.
	fmt.Printf("Площадь: %.1f\n", C.area(r))

	// Передаём поле-структуру origin в C-функцию,
	// которая ожидает полный Rect — нельзя, но можно
	// передать указатель на весь Rect для мутации поля.
	fmt.Printf("До scale_width: size.x = %.1f\n", r.size.x)
	C.scale_width(&r, 3.0)
	fmt.Printf("После scale_width(3.0): size.x = %.1f\n", r.size.x)

	// Можно читать и записывать отдельные поля напрямую.
	r.origin.x = 100
	r.origin.y = 200
	fmt.Printf("Новый origin: (%.0f, %.0f)\n", r.origin.x, r.origin.y)
	fmt.Printf("Площадь после масштабирования: %.1f\n", C.area(r))
}

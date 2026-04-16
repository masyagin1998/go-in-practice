package main

// Пример вызова C-функции из Go.
// C-код находится в greet.c / greet.h.

/*
#include "greet.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	// Создаём C-строку из Go-строки.
	name := C.CString("World")
	defer C.free(unsafe.Pointer(name))

	// Вызываем C-функцию greet.
	C.greet(name)

	fmt.Println("Done (from Go)")
}

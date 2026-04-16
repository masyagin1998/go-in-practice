package bench

// Обёртки для вызова пустой C-функции.

/*
#include "empty.h"
*/
import "C"
import "unsafe"

// CgoEmpty — вызов пустой C-функции через стандартный CGo.
func CgoEmpty(x int) {
	C.empty(C.int(x))
}

// CEmptyPtr — указатель на C-функцию empty для fastcgo.
var CEmptyPtr = unsafe.Pointer(C.empty)

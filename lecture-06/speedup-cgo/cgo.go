package speedup

// Обёртки для вызова функции Аккермана на C.

/*
#cgo CFLAGS: -O3
#include "ackermann.h"
*/
import "C"
import "unsafe"

// CgoAckermann — вызов через стандартный CGo.
func CgoAckermann(m, n int) int {
	return int(C.ackermann(C.int(m), C.int(n)))
}

// CAckermannPtr — указатель на C-функцию для fastcgo.
var CAckermannPtr = unsafe.Pointer(C.ackermann)

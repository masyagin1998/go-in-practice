package main

// CGo НЕ поддерживает вызов variadic C-функций (printf, sprintf, ...).
//
// printf объявлен как: int printf(const char *format, ...)
// CGo генерирует обёртку только для фиксированных аргументов.
// Передача variadic-аргументов — ошибка компиляции.
//
// Решение: написать C-обёртку с фиксированными параметрами.

/*
#include <stdio.h>
*/
import "C"

func main() {
	// Не компилируется:
	//   cannot use _Cgo_use (type func(interface {})) as type func(interface {})
	//   too many arguments in call to _Cfunc_printf
	C.printf(C.CString("answer = %d\n"), C.int(42))
}

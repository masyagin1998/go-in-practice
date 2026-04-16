package main

// Два способа передачи строк из Go в C:
//
// 1. C.CString — копирует Go-строку в C-heap (malloc), добавляет '\0'.
//    Нужно вручную вызывать C.free(). Безопасно: C владеет памятью.
//
// 2. Передача указателя + длины напрямую (без копирования).
//    Go-строка не содержит нуль-терминатора — C-код должен использовать длину.
//    Также можно использовать встроенный CGo-тип _GoString_
//    (доступен только в inline-C, не в отдельных .c файлах).

/*
#include "strutil.h"
#include <stdlib.h>
#include <stdio.h>

// print_gostring использует встроенный CGo-тип _GoString_.
// _GoString_ доступен только в CGo-preamble, не в отдельных .c файлах.
// Go автоматически конвертирует string → _GoString_ при вызове.
static void print_gostring(_GoString_ s) {
    size_t      n = _GoStringLen(s);
    const char* p = _GoStringPtr(s);
    printf("[C] _GoString_ (len=%zu): \"%.*s\"\n", n, (int)n, p);
    fflush(stdout);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	goStr := "Привет, CGo!"

	// === Способ 1: C.CString ===
	// Выделяет память в C-heap через malloc, копирует байты, добавляет '\0'.
	cStr := C.CString(goStr)
	defer C.free(unsafe.Pointer(cStr))

	fmt.Println("--- C.CString (копия в C-heap, с '\\0') ---")
	C.print_cstring(cStr)

	// C.CString возвращает *C.char — можно мутировать в C.
	C.to_upper_cstring(cStr)
	// Конвертируем обратно в Go через C.GoString (ещё одна копия).
	upper := C.GoString(cStr)
	fmt.Printf("[Go] После to_upper: %q\n", upper)

	// === Способ 2: указатель + длина (без копирования) ===
	fmt.Println("\n--- Указатель + длина (без копирования, без '\\0') ---")
	C.print_raw_gostring(
		(*C.char)(unsafe.Pointer(unsafe.StringData(goStr))),
		C.size_t(len(goStr)),
	)

	// === Способ 3: _GoString_ (inline CGo) ===
	// Функция с параметром _GoString_ принимает Go string напрямую.
	fmt.Println("\n--- _GoString_ (встроенный CGo-тип) ---")
	C.print_gostring(goStr)

	// Работает и с подстроками.
	C.print_gostring("Подстрока")
}

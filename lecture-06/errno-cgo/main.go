package main

// Обработка ошибок C через errno в CGo.
//
// В C нет исключений — ошибки сигнализируются через:
//   1. Возвращаемое значение (обычно -1 или NULL)
//   2. Глобальную переменную errno (целое число, код ошибки)
//
// CGo поддерживает специальный синтаксис: если C-функция возвращает
// одно значение, можно получить ДВА значения при вызове из Go:
//
//   result, err := C.some_func(args)
//
// Здесь err — это errno, обёрнутый в тип error (syscall.Errno).
// CGo всегда возвращает errno, даже если функция завершилась успешно —
// в этом случае err == nil только если errno == 0.
//
// ВАЖНО: errno — это thread-local переменная. CGo корректно читает её
// сразу после вызова C-функции, до переключения горутины.

/*
#include "fileops.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	// === Пример 1: открытие несуществующего файла ===
	// open() вернёт -1 и установит errno = ENOENT (No such file or directory).
	fmt.Println("=== open() несуществующего файла ===")
	path := C.CString("/tmp/этого_файла_точно_нет_12345")
	defer C.free(unsafe.Pointer(path))

	fd, err := C.try_open(path)
	fmt.Printf("fd = %d, err = %v\n", fd, err)
	if err != nil {
		fmt.Printf("Тип ошибки: %T\n", err)
		// err — это syscall.Errno, можно сравнивать с константами.
	}

	// === Пример 2: открытие существующего файла (errno == 0 → err == nil) ===
	fmt.Println("\n=== open() существующего файла ===")
	devNull := C.CString("/dev/null")
	defer C.free(unsafe.Pointer(devNull))

	fd2, err2 := C.try_open(devNull)
	fmt.Printf("fd = %d, err = %v\n", fd2, err2)
	if err2 != nil {
		fmt.Printf("Тип ошибки: %T\n", err2)
	}

	// === Пример 3: chdir в несуществующую директорию ===
	// chdir() вернёт -1 и установит errno = ENOENT.
	fmt.Println("\n=== chdir() в несуществующую директорию ===")
	badDir := C.CString("/несуществующий/путь")
	defer C.free(unsafe.Pointer(badDir))

	ret, err3 := C.try_chdir(badDir)
	fmt.Printf("ret = %d, err = %v\n", ret, err3)

	// === Пример 4: пользовательская функция с errno ===
	// safe_div устанавливает errno = EINVAL при делении на ноль.
	fmt.Println("\n=== safe_div: деление на ноль (errno = EINVAL) ===")
	result, err4 := C.safe_div(42, 0)
	fmt.Printf("42 / 0: result = %d, err = %v\n", result, err4)

	fmt.Println("\n=== safe_div: успешное деление ===")
	result2, err5 := C.safe_div(42, 7)
	fmt.Printf("42 / 7: result = %d, err = %v\n", result2, err5)
}

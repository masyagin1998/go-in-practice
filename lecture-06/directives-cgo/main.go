package main

// Демонстрация директив #cgo — управление флагами компиляции и линковки.
//
// Директивы #cgo пишутся в преамбуле (перед import "C") и задают параметры
// сборки C-кода. Формат:
//
//   #cgo [УСЛОВИЕ] ПЕРЕМЕННАЯ: ЗНАЧЕНИЕ
//
// Основные переменные:
//   CFLAGS    — флаги компилятора C (gcc/clang), напр. -I, -D, -O2, -Wall
//   CXXFLAGS  — флаги компилятора C++ (g++/clang++)
//   LDFLAGS   — флаги линковщика, напр. -L (путь к .so/.a) и -l (имя библиотеки)
//   pkg-config — автоматическое определение флагов через утилиту pkg-config
//
// Условия (опционально):
//   #cgo linux       — применять только на Linux
//   #cgo darwin      — применять только на macOS
//   #cgo windows     — применять только на Windows
//   #cgo amd64       — применять только на amd64
//   #cgo linux,amd64 — конъюнкция: Linux И amd64
//   #cgo !windows    — отрицание: все, кроме Windows
//
// В этом примере удаление директивы ломает сборку:
//
//   -Iinclude          — заголовки лежат в include/, без -I → fatal error: No such file
//   -DPLATFORM_LINUX   — platform.c использует #ifdef PLATFORM_LINUX, без -D → #error
//   -lm                — линковка libm для sqrt(). На современном Linux/gcc линкуется
//                        неявно, но на macOS и при cross-compile обязательна.

/*
#cgo CFLAGS: -Iinclude -Wall -Wextra -Wno-unused-parameter -O2
#cgo LDFLAGS: -lm

#cgo linux  CFLAGS: -DPLATFORM_LINUX
#cgo darwin CFLAGS: -DPLATFORM_DARWIN

#include "mathutil.h"
#include "platform.h"
*/
import "C"

import "fmt"

func main() {
	// === CFLAGS: -Iinclude ===
	// Заголовки mathutil.h и platform.h лежат в include/.
	// Без -Iinclude: fatal error: mathutil.h: No such file or directory
	fmt.Println("=== CFLAGS: -Iinclude (путь к заголовкам) ===")
	fmt.Println("  #include \"mathutil.h\" → ищет в include/mathutil.h")

	// === LDFLAGS: -lm ===
	// mathutil.c использует sqrt() из libm.
	// На современном Linux/gcc libm линкуется неявно, но на macOS и при
	// cross-compile без -lm будет: undefined reference to `sqrt'.
	fmt.Println("\n=== LDFLAGS: -lm (линковка libm для sqrt) ===")
	h := C.hypotenuse(3.0, 4.0)
	fmt.Printf("  hypotenuse(3, 4) = %.1f\n", h)

	// === Платформозависимые CFLAGS: -DPLATFORM_LINUX ===
	// platform.c: #if defined(PLATFORM_LINUX) ... #else #error
	// Без #cgo linux CFLAGS: -DPLATFORM_LINUX → ошибка компиляции.
	fmt.Println("\n=== Платформозависимые CFLAGS: -DPLATFORM_LINUX ===")
	name := C.GoString(C.platform_name())
	ps := C.page_size()
	fmt.Printf("  Платформа: %s\n", name)
	fmt.Printf("  Размер страницы: %d байт\n", ps)
}

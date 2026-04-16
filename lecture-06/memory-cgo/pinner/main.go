package main

// runtime.Pinner (Go 1.21+) позволяет «закрепить» Go-объект в памяти,
// чтобы GC не перемещал его, пока C-код хранит указатель.
//
// В этом примере C-код сохраняет Go-указатель в глобальной переменной
// через buffer_set(). После этого buffer_print() и buffer_scale()
// обращаются к Go-памяти напрямую — без повторной передачи указателя.
//
// Между каждым CGo-вызовом мы явно запускаем GC + debug.FreeOSMemory(),
// чтобы максимально спровоцировать перемещение объектов.
// Без Pinner это UB: GC может переместить массив, и глобальный указатель
// в C будет указывать на мусор.

/*
#include "buffer.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"unsafe"
)

// aggressiveGC запускает сборку мусора и возвращает память ОС.
// Цель — максимально спровоцировать перемещение/уплотнение heap.
func aggressiveGC() {
	runtime.GC()
	debug.FreeOSMemory()
	fmt.Println("[Go] runtime.GC() + debug.FreeOSMemory()")
}

func main() {
	// Аллоцируем слайс в Go.
	data := make([]C.int, 5)
	for i := range data {
		data[i] = C.int(i + 1) // [1, 2, 3, 4, 5]
	}

	// Закрепляем underlying array слайса через Pinner.
	// После Pin() GC не будет перемещать этот массив.
	var pinner runtime.Pinner
	pinner.Pin(&data[0])
	defer pinner.Unpin() // Снимаем закрепление при выходе.

	// Передаём указатель на Go-память в C. C сохраняет его в глобальной переменной.
	C.buffer_set((*C.int)(unsafe.Pointer(&data[0])), C.size_t(len(data)))

	aggressiveGC() // GC между buffer_set и buffer_print — указатель в C всё ещё валиден?

	// Последующие вызовы НЕ получают указатель — они читают из глобальной
	// переменной C, которая указывает на Go-память.
	fmt.Println("\n=== Исходные данные (C читает из глобального указателя) ===")
	C.buffer_print()

	aggressiveGC() // Ещё раз — между print и scale.

	// C мутирует Go-память через глобальный указатель.
	fmt.Println("\n=== buffer_scale(10) — C мутирует Go-массив ===")
	C.buffer_scale(10)

	aggressiveGC() // И ещё раз — после мутации, перед чтением.

	fmt.Println("\n=== Проверяем: данные на месте? ===")
	C.buffer_print()

	// Go видит изменения — это одна и та же память.
	fmt.Printf("\n[Go] слайс после мутации в C: %v\n", data)
}

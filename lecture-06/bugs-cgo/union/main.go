package main

// C union в Go: CGo представляет union как массив байт [N]byte.
//
// В C:
//   union Value { int i; float f; char bytes[4]; };
//
// В Go:
//   type Value = [4]byte    // все поля слиты в массив байт
//
// Доступа к t.val.i или t.val.f из Go нет.
//
// Два способа интерпретировать:
//   1. C-аксессоры (tagged_as_int, tagged_as_float) — безопасно
//   2. unsafe.Pointer + приведение типа — ручной доступ, без CGo-вызова

/*
#include "value.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

func main() {
	ti := C.make_int(42)
	tf := C.make_float(3.14)

	// C работает с union напрямую.
	C.tagged_print(ti)
	C.tagged_print(tf)

	// Go видит val как [4]byte.
	fmt.Printf("\n[Go] sizeof(C.Value)  = %d\n", unsafe.Sizeof(C.Value{}))
	fmt.Printf("[Go] sizeof(C.Tagged) = %d\n", unsafe.Sizeof(C.Tagged{}))
	fmt.Printf("[Go] ti.val = %v (тип %T) — сырые байты\n", ti.val, ti.val)
	fmt.Printf("[Go] tf.val = %v (тип %T) — сырые байты\n", tf.val, tf.val)

	// ti.val.i → ошибка компиляции: [4]byte has no field i

	// === Способ 1: C-аксессоры ===
	fmt.Println("\n=== Способ 1: C-аксессоры ===")
	fmt.Printf("  ti → int:   %d\n", C.tagged_as_int(ti))
	fmt.Printf("  tf → float: %.2f\n", C.tagged_as_float(tf))

	// === Способ 2: unsafe.Pointer ===
	// Интерпретируем [4]byte как int32 или float32 через unsafe.
	fmt.Println("\n=== Способ 2: unsafe.Pointer ===")

	iVal := *(*int32)(unsafe.Pointer(&ti.val[0]))
	fmt.Printf("  ti → int:   %d\n", iVal)

	fVal := *(*float32)(unsafe.Pointer(&tf.val[0]))
	fmt.Printf("  tf → float: %.2f\n", fVal)

	// === Способ 3: encoding/binary ===
	// Без unsafe — через байтовую интерпретацию.
	fmt.Println("\n=== Способ 3: encoding/binary (без unsafe) ===")

	iBits := binary.LittleEndian.Uint32(ti.val[:])
	fmt.Printf("  ti → int:   %d\n", int32(iBits))

	fBits := binary.LittleEndian.Uint32(tf.val[:])
	fmt.Printf("  tf → float: %.2f\n", math.Float32frombits(fBits))
}

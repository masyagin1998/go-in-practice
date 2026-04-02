// go run main.go
//
// Аналог C-структуры С __packed__:
//
//   struct __attribute__((__packed__)) Shared {
//       int32_t shit;   // offset 0,  size 4
//       int64_t value;  // offset 4,  size 8
//   };  // sizeof = 12
//
// В Go нет __packed__, поэтому эмулируем через unsafe.Pointer.
// value на смещении 4 — невыровнен по 8, но всё ещё внутри
// одной кэш-линии (байты 4..11 из 64) → на x86-64 torn read
// НЕ возникает, хотя мы нарушили выравнивание.

package main

import (
	"cpu-atomic/torntest"
	"fmt"
	"unsafe"
)

var buf [256]byte

func main() {
	// struct __packed__ { int32_t shit; int64_t value; }
	// sizeof = 12, value at offset 4
	aligned := torntest.AlignTo64(uintptr(unsafe.Pointer(&buf[0])))
	ptr := torntest.PtrAtOffset(aligned, 4) // offset 4: после int32 без padding

	fmt.Printf("Эмулируем packed struct: sizeof=12, value at offset 4\n")
	fmt.Printf("Address of value: %p (offset mod 64 = %d)\n",
		ptr, uintptr(unsafe.Pointer(ptr))%64)

	torntest.Run(ptr)
}

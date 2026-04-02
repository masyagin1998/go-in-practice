// go run main.go
//
// Аналог C-структуры с разрывом кэш-линии:
//
//   struct __attribute__((__packed__)) Shared {
//       char shit[60];          // offset 0,  size 60
//       volatile int64_t value; // offset 60, size 8
//   } __attribute__((aligned(64)));
//
// value на смещении 60 — байты [60..63] в одной кэш-линии,
// [64..67] в следующей → torn read ВОЗМОЖЕН.

package main

import (
	"cpu-atomic/torntest"
	"fmt"
	"unsafe"
)

var buf [256]byte

func main() {
	// struct __packed__ { char shit[60]; int64_t value; } __aligned__(64)
	// sizeof = 68, value at offset 60 — пересекает границу кэш-линии
	aligned := torntest.AlignTo64(uintptr(unsafe.Pointer(&buf[0])))
	ptr := torntest.PtrAtOffset(aligned, 60) // offset 60: разрыв на границе линий

	fmt.Printf("Эмулируем packed struct: sizeof=68, value at offset 60\n")
	fmt.Printf("Address of value: %p (offset mod 64 = %d)\n",
		ptr, uintptr(unsafe.Pointer(ptr))%64)

	torntest.Run(ptr)
}

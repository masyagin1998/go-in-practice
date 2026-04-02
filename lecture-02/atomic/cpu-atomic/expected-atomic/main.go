// go run main.go
//
// Аналог C-структуры БЕЗ __packed__:
//
//   struct Shared {
//       int32_t shit;   // offset 0,  size 4
//       // 4 байта padding
//       int64_t value;  // offset 8,  size 8
//   };  // sizeof = 16
//
// Go делает то же самое: выравнивает int64 по 8 байтам.
// value целиком внутри одной кэш-линии → torn read невозможен.

package main

import (
	"cpu-atomic/torntest"
	"fmt"
	"unsafe"
)

type Shared struct {
	shit  int32
	value int64
}

var sharedData Shared

func main() {
	fmt.Printf("sizeof(Shared) = %d\n", unsafe.Sizeof(sharedData))
	fmt.Printf("offsetof(value) = %d\n", unsafe.Offsetof(sharedData.value))
	fmt.Printf("Address of value: %p (offset mod 64 = %d)\n",
		&sharedData.value,
		uintptr(unsafe.Pointer(&sharedData.value))%64)

	torntest.Run(&sharedData.value)
}

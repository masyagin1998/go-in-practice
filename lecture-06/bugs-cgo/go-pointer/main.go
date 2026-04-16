package main

// Передача Go-указателя, содержащего другие Go-указатели, в C.
//
// Правило CGo: Go-код может передать Go-указатель в C, только если
// Go-память, на которую он указывает, НЕ содержит других Go-указателей.
//
// Слайс []int — это структура {ptr, len, cap}, где ptr — Go-указатель.
// Если передать &slice в C, то C получит указатель на Go-память,
// внутри которой лежит ещё один Go-указатель (на underlying array).
//
// cgocheck (включён по умолчанию) обнаруживает это и паникует:
//   runtime error: cgo argument has Go pointer to unpinned Go pointer

/*
#include "store.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Data содержит слайс — внутри слайса есть Go-указатель на underlying array.
type Data struct {
	Values []int
}

func main() {
	d := &Data{
		Values: []int{1, 2, 3, 4, 5},
	}

	fmt.Printf("[Go] &d = %p, d.Values underlying = %p\n",
		unsafe.Pointer(d), unsafe.Pointer(&d.Values[0]))

	// Паника: cgo argument has Go pointer to unpinned Go pointer.
	// Go-структура Data содержит слайс, а слайс — это Go-указатель.
	C.store_ptr(unsafe.Pointer(d))

	C.print_ptr()
}

package main

// Передача простых значений из Go в C.
// Go-типы (C.int, C.double) — это просто typedef'ы,
// значения копируются по стеку, аллокаций нет.

/*
#include "math.h"
*/
import "C"

import "fmt"

func main() {
	// C.int, C.double — псевдотипы CGo, копируются по значению.
	sum := C.add_ints(10, 32)
	fmt.Printf("add_ints(10, 32)       = %d (тип %T)\n", sum, sum)

	prod := C.multiply_doubles(3.14, 2.0)
	fmt.Printf("multiply_doubles(3.14, 2.0) = %f\n", prod)

	// bool в C (stdbool.h) маппится на C._Bool.
	fmt.Printf("is_positive(5)  = %v\n", C.is_positive(5))
	fmt.Printf("is_positive(-3) = %v\n", C.is_positive(-3))
}

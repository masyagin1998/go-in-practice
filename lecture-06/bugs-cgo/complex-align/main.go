package main

// Packed-структура с complex float: CGo не может представить
// поля с нарушенным alignment.
//
// В C:
//   struct __attribute__((packed)) {
//       char          tag;    // offset 0
//       float complex val;    // offset 1 (alignment нарушен!)
//       int           count;  // offset 9
//   };
//
// В Go: CGo сворачивает val и count в непрозрачный массив байт:
//   type Sample struct {
//       tag C.char
//       _   [12]byte   // val + count — доступа к ним нет!
//   }
//
// Обращение к s.val или s.count — ошибка компиляции.
//
// Решение: C-функции-аксессоры (sample_re, sample_im, sample_count).

/*
#include "sample.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	s := C.make_sample('A', 3.0, 4.0, 42)

	// C видит все поля нормально.
	C.sample_print(s)

	// Go видит только tag — остальное спрятано в _:[12]byte.
	fmt.Printf("\n[Go] sizeof(C.Sample) = %d (C: 13)\n", unsafe.Sizeof(s))
	fmt.Printf("[Go] s.tag = '%c'\n", s.tag)

	// s.val   → ошибка компиляции: no field or method val
	// s.count → ошибка компиляции: no field or method count

	// Workaround: C-аксессоры.
	fmt.Printf("\n[Go] Через C-аксессоры:\n")
	fmt.Printf("  tag   = '%c'\n", C.sample_tag(s))
	fmt.Printf("  re    = %.1f\n", C.sample_re(s))
	fmt.Printf("  im    = %.1f\n", C.sample_im(s))
	fmt.Printf("  count = %d\n", C.sample_count(s))

	// Для наглядности: как Go видит структуру.
	fmt.Printf("\n[Go] Сырое содержимое: %+v\n", s)
	fmt.Println("     (tag виден, остальные поля — непрозрачный массив _)")
}

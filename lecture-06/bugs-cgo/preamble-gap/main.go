package main

// Пустая строка между комментарием-преамбулой и import "C"
// приводит к тому, что CGo молча игнорирует преамбулу.
//
// Преамбула (код между /* */ перед import "C") должна идти
// НЕПОСРЕДСТВЕННО перед import "C" — без пустых строк.
//
// С пустой строкой комментарий становится обычным Go-комментарием,
// а не CGo-преамбулой. #include не выполняется → C-функции недоступны.

/*
#include <stdio.h>

static void hello() {
    printf("Hello from C!\n");
    fflush(stdout);
}
*/

import "C"

func main() {
	// Не компилируется: could not determine kind of name for C.hello
	C.hello()
}

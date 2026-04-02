// Вставить в godbolt.org (компилятор: Go, флаги: -gcflags=-S)
// Сравнить ассемблер с atomic версией.

package main

var counter int64

func Inc() {
	counter++
}

func main() { Inc() }

// В ассемблере будет:
//   INCQ  main.counter(SB)   // read-modify-write БЕЗ LOCK
//
// INCQ на памяти — одна инструкция, но НЕ атомарная:
// CPU внутри делает load → inc → store, и между ними
// другое ядро может вклиниться. Нет LOCK префикса.

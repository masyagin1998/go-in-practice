// Вставить в godbolt.org (компилятор: x86-64 gc 1.21+)
// Сравнить ассемблер с atomic версией (godbolt-atomic/xchg.go).
//
// "Спинлок" без atomic — сломанный.
// Компилятор генерирует обычные MOV вместо XCHG.

package main

var flag int32

func TryLock() bool {
	old := flag
	flag = 1
	return old == 0
}

func Unlock() {
	flag = 0
}

func main() {
	TryLock()
	Unlock()
}

// TryLock в ассемблере:
//   MOVL  flag(SB), AX     // load (обычный, без LOCK)
//   MOVL  $1, flag(SB)     // store (обычный, без LOCK)
//   TESTL AX, AX           // old == 0?
//
// Два отдельных MOV вместо одного XCHG.
// Между load и store другое ядро может:
//   1. Тоже прочитать flag=0
//   2. Тоже записать flag=1
//   → оба "захватили" лок. Mutual exclusion сломан.

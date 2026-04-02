// Вставить в godbolt.org (компилятор: x86-64 gc 1.21+)
// Сравнить ассемблер с non-atomic версией.

package main

import "sync/atomic"

var counter atomic.Int64

func Inc() {
	counter.Add(1)
}

func main() { Inc() }

// В ассемблере будет LOCK XADDQ — атомарный инкремент.
// LOCK префикс блокирует шину памяти (или кэш-линию),
// гарантируя атомарность на уровне CPU.

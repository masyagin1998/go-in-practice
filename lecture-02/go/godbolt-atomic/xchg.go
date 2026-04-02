// Вставить в godbolt.org (компилятор: x86-64 gc 1.21+)
// Генерирует XCHG — атомарный обмен значениями.

package main

import "sync/atomic"

var flag int32

// TryLock — атомарно меняет flag: 0→1.
// Если старое значение 0 — мы захватили лок.
// atomic.SwapInt32 компилируется в XCHGL.
// XCHG на x86 всегда имеет неявный LOCK — даже без префикса.
func TryLock() bool {
	old := atomic.SwapInt32(&flag, 1)
	return old == 0
}

func Unlock() {
	atomic.StoreInt32(&flag, 0)
}

func main() {
	TryLock()
	Unlock()
}

// TryLock в ассемблере:
//   MOVL  $1, AX
//   XCHGL AX, flag(SB)     // атомарно: AX ↔ flag
//   TESTL AX, AX           // old == 0?
//
// XCHG на x86 ВСЕГДА атомарен на памяти — LOCK подразумевается.
// Это единственная инструкция с неявным LOCK.
// Именно поэтому простейший spinlock (test-and-set) строится на XCHG.

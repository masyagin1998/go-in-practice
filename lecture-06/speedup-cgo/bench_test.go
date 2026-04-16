package speedup

// Сравнение производительности функции Аккермана:
// 1. Pure Go
// 2. C через CGo
// 3. C через fastcgo
//
// Демонстрирует, что даже с overhead'ом CGo/fastcgo вызов C-функции
// может быть быстрее, чем аналогичная реализация на Go, за счёт
// более агрессивных оптимизаций компилятора C (gcc -O2).

import (
	"testing"

	"github.com/nitrix/fastcgo"
)

// Ackermann — функция Аккермана на чистом Go.
func Ackermann(m, n int) int {
	if m == 0 {
		return n + 1
	}
	if n == 0 {
		return Ackermann(m-1, 1)
	}
	return Ackermann(m-1, Ackermann(m, n-1))
}

// BenchmarkGoAckermann — чистый Go.
func BenchmarkGoAckermann(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Ackermann(3, 12)
	}
}

// BenchmarkCgoAckermann — C через стандартный CGo.
func BenchmarkCgoAckermann(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CgoAckermann(3, 12)
	}
}

// BenchmarkFastcgoAckermann — C через fastcgo (только подмена стека).
func BenchmarkFastcgoAckermann(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fastcgo.UnsafeCall2Return1(CAckermannPtr, 3, 12)
	}
}

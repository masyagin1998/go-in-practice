package bench

// Сравнение стоимости вызова:
// 1. Пустая Go-функция (//go:noinline)
// 2. Пустая C-функция через CGo
// 3. Пустая C-функция через fastcgo (подмена стека, без накладных расходов runtime)

import (
	"testing"

	"github.com/nitrix/fastcgo"
)

//go:noinline
func emptyGo(x int) {}

// BenchmarkGoCall — вызов пустой Go-функции.
func BenchmarkGoCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		emptyGo(42)
	}
}

// BenchmarkCgoCall — вызов пустой C-функции через стандартный CGo.
func BenchmarkCgoCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CgoEmpty(42)
	}
}

// BenchmarkFastcgoCall — вызов пустой C-функции через fastcgo (только подмена стека).
func BenchmarkFastcgoCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fastcgo.UnsafeCall1(CEmptyPtr, 42)
	}
}

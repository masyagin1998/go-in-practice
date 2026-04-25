package bench

import (
	"testing"
)

var Sink int

func genInput(n int) []int {
	xs := make([]int, n)
	for i := range xs {
		xs[i] = i
	}
	return xs
}

func TestSumIndex(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5}
	if got := sumIndex(xs); got != 15 {
		t.Fatalf("sumIndex = %d, want 15", got)
	}
	if got := sumIndex(nil); got != 0 {
		t.Fatalf("sumIndex(nil) = %d, want 0", got)
	}
}

func TestSumRange(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5}
	if got := sumRange(xs); got != 15 {
		t.Fatalf("sumRange = %d, want 15", got)
	}
	if got := sumRange(nil); got != 0 {
		t.Fatalf("sumRange(nil) = %d, want 0", got)
	}
}

func TestSumUnroll(t *testing.T) {
	for _, n := range []int{0, 1, 3, 4, 5, 7, 8, 100} {
		xs := genInput(n)
		want := n * (n - 1) / 2
		if got := sumUnroll(xs); got != want {
			t.Fatalf("sumUnroll(n=%d) = %d, want %d", n, got, want)
		}
	}
}

func BenchmarkSum(b *testing.B) {
	sizes := []int{1_000, 100_000, 1_000_000}
	impls := []struct {
		name string
		fn   func([]int) int
	}{
		{"index", sumIndex},
		{"range", sumRange},
		{"unroll", sumUnroll},
	}

	for _, n := range sizes {
		xs := genInput(n)

		for _, impl := range impls {
			b.Run("size="+itoa(n)+"/"+impl.name, func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					Sink = impl.fn(xs)
				}
			})
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	const digits = "0123456789"
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	return string(buf[i:])
}

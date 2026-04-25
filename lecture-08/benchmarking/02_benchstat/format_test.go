package parse

import "testing"

var Sink string

func TestFormatFmt(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-42, "-42"},
		{1234567, "1234567"},
	}
	for _, c := range cases {
		if got := FormatFmt(c.in); got != c.want {
			t.Errorf("FormatFmt(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatStrconv(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-42, "-42"},
		{1234567, "1234567"},
	}
	for _, c := range cases {
		if got := FormatStrconv(c.in); got != c.want {
			t.Errorf("FormatStrconv(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatEqual(t *testing.T) {
	for _, n := range []int{0, 1, -1, 100, -100, 1<<31 - 1, -(1 << 31)} {
		if FormatFmt(n) != FormatStrconv(n) {
			t.Errorf("mismatch for n=%d: fmt=%q strconv=%q",
				n, FormatFmt(n), FormatStrconv(n))
		}
	}
}

func BenchmarkFormatFmt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Sink = FormatFmt(i)
	}
}

func BenchmarkFormatStrconv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Sink = FormatStrconv(i)
	}
}

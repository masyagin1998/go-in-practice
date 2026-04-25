package bench

import (
	"testing"
)

var SinkBuf []byte

func BenchmarkAllocBufNaive(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 1024)
		buf[0] = byte(i)
		SinkBuf = buf
	}
}

func BenchmarkAllocBufReused(b *testing.B) {
	b.ReportAllocs()
	buf := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf[0] = byte(i)
		SinkBuf = buf
	}
}

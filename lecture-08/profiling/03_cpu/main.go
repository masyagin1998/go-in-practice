package main

// CPU-профиль: классический случай — наивная склейка строк через +=
// против strings.Builder. Снимем cpu.pprof и посмотрим, где именно
// горит CPU.

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
)

// joinNaive — O(N^2) по времени: каждое += копирует всю строку.
// Да, vet/staticcheck ругается — это и есть смысл примера.
func joinNaive(parts []string) string {
	s := ""
	for _, p := range parts {
		s += p //nolint:stringsbuilder // намеренно: показываем плохой паттерн
	}
	return s
}

// joinBuilder — O(N): Builder переиспользует буфер.
func joinBuilder(parts []string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

func main() {
	parts := make([]string, 50_000)
	for i := range parts {
		parts[i] = "hello"
	}

	f, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 1 кГц вместо дефолтных 100 Гц: без этого joinBuilder
	// (десятки µs за прогон) не наберёт ни одного сэмпла.
	// runtime.SetCPUProfileRate(1000)
	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	// joinNaive — O(N^2), ему хватает 3 прогонов, чтобы сжечь секунды CPU.
	var r1, r2 string
	for i := 0; i < 3; i++ {
		r1 = joinNaive(parts)
	}
	// joinBuilder — O(N), один прогон ~десятки µs. Гоняем много раз,
	// чтобы он проявился в flame graph как отдельный узел.
	for i := 0; i < 3; i++ {
		r2 = joinBuilder(parts)
	}
	fmt.Printf("naive=%d symbols, builder=%d symbols\n", len(r1), len(r2))
}

package main

// Heap-профиль: три точки аллокации с разной судьбой.
//
//   1. leakBig    — крупные буферы навсегда в глобал
//                   → видны и в inuse, и в alloc (живут до конца).
//   2. churnSmall — мелкие буферы сразу теряются (escape через sink,
//                   на след. итерации старый умирает)
//                   → доминируют в alloc, в inuse ≈ 0.
//   3. cacheHalf  — средние буферы, половину обнуляем
//                   → inuse в 2 раза меньше alloc.
//
// //go:noinline разводит функции по отдельным фреймам стека —
// иначе компилятор склеит их с main, и в pprof будет один узел.
// На flamegraph должно получиться три чётко различимых «башни».

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	leakStore  [][]byte // 1: держим вечно
	cacheStore [][]byte // 3: держим только половину
	sink       []byte   // 2: escape-trick — гонит make() в heap
)

//go:noinline
func leakBig() {
	leakStore = append(leakStore, make([]byte, 64*1024))
}

//go:noinline
func churnSmall() {
	// присваивание глобалу заставляет escape analysis уйти в heap.
	// На следующем вызове sink перезапишется, прошлый буфер умрёт.
	sink = make([]byte, 4*1024)
}

//go:noinline
func cacheHalf() {
	cacheStore = append(cacheStore, make([]byte, 16*1024))
}

func main() {
	// 1) утечка: 1000 × 64 KiB ≈ 64 MiB и в inuse, и в alloc
	for range 1_000 {
		leakBig()
	}

	// 2) churn: 100k × 4 KiB ≈ 400 MiB только в alloc, в inuse ≈ 0
	for range 100_000 {
		churnSmall()
	}
	sink = nil // отпускаем последний буфер, иначе 4 KiB останется в inuse

	// 3) кеш с половинной очисткой: 4000 × 16 KiB ≈ 64 MiB alloc → 32 MiB inuse
	for range 4_000 {
		cacheHalf()
	}
	for i := range len(cacheStore) / 2 {
		cacheStore[i] = nil
	}

	// runtime.GC() обязателен: без него churn-мусор и обнулённая половина
	// кеша всё ещё считаются "живыми" в inuse.
	runtime.GC()

	f, err := os.Create("heap.pprof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}

	live := 0
	for _, b := range cacheStore {
		if b != nil {
			live++
		}
	}
	fmt.Printf("leakStore: %d буферов держим\n", len(leakStore))
	fmt.Printf("cacheStore: %d живых из %d\n", live, len(cacheStore))
	fmt.Printf("churnSmall: %d аллокаций (все мертвы)\n", 100_000)
	fmt.Println()
	fmt.Println("inuse:  go tool pprof -top -sample_index=inuse_space heap.pprof")
	fmt.Println("alloc:  go tool pprof -top -sample_index=alloc_space heap.pprof")
	fmt.Println("flame:  go tool pprof -http=: heap.pprof")
}

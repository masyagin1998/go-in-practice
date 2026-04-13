// Пример 4: Tiny-аллокатор Go.
//
// Для объектов < 16 байт БЕЗ указателей Go использует специальный tiny-аллокатор:
// несколько мелких объектов упаковываются в один 16-байтный слот span'а.
// Это экономит ~12% аллокаций и ~20% памяти на типичных нагрузках.
//
// Условия для tiny-аллокатора:
//   - Размер < 16 байт
//   - Объект НЕ содержит указателей (noscan)
//
// Упаковка видна через HeapObjects (реальные объекты в хипе),
// а не через Mallocs (каждый вызов mallocgc считается отдельно).
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

func main() {
	const N = 100_000

	// =============================================================
	// Тест 1: Мелкие объекты без указателей (tiny allocator)
	// =============================================================
	fmt.Println("=== Tiny-аллокатор: int8 (1 байт, без указателей) ===")

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	ptrs := make([]*int8, N)
	for i := range ptrs {
		v := int8(i)
		ptrs[i] = &v
	}

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	heapObjDelta := after.HeapObjects - before.HeapObjects
	heapAllocDelta := after.HeapAlloc - before.HeapAlloc
	fmt.Printf("  Создано объектов:     %d\n", N)
	fmt.Printf("  HeapObjects прирост:  %d (реальных объектов в хипе)\n", heapObjDelta)
	fmt.Printf("  HeapAlloc прирост:    %d байт\n", heapAllocDelta)
	fmt.Printf("  Упаковка:             %.1fx (логических на 1 heap-объект)\n",
		float64(N)/float64(heapObjDelta))
	fmt.Printf("  Байт на объект:       %.1f (sizeof=%d, в 16-байтный слот влезает %d)\n",
		float64(heapAllocDelta)/float64(N),
		unsafe.Sizeof(int8(0)),
		16/unsafe.Sizeof(int8(0)))
	runtime.KeepAlive(ptrs)

	// =============================================================
	// Тест 2: Мелкие объекты С указателями (обычный аллокатор)
	// =============================================================
	fmt.Println("\n=== Обычный аллокатор: структура с указателем ===")

	runtime.GC()
	runtime.ReadMemStats(&before)

	type withPtr struct {
		val int8
		ptr *int8
	}
	ptrs2 := make([]*withPtr, N)
	for i := range ptrs2 {
		v := int8(i)
		ptrs2[i] = &withPtr{val: v, ptr: &v}
	}

	runtime.ReadMemStats(&after)

	heapObjDelta = after.HeapObjects - before.HeapObjects
	heapAllocDelta = after.HeapAlloc - before.HeapAlloc
	fmt.Printf("  Создано объектов:     %d\n", N)
	fmt.Printf("  HeapObjects прирост:  %d\n", heapObjDelta)
	fmt.Printf("  HeapAlloc прирост:    %d байт\n", heapAllocDelta)
	fmt.Printf("  Упаковка:             %.1fx\n",
		float64(N)/float64(heapObjDelta))
	fmt.Printf("  sizeof(withPtr) = %d (содержит указатель → tiny НЕ работает)\n",
		unsafe.Sizeof(withPtr{}))
	runtime.KeepAlive(ptrs2)

	// =============================================================
	// Тест 3: Сравнение разных размеров без указателей
	// =============================================================
	fmt.Println("\n=== Зависимость упаковки от размера (без указателей) ===")
	fmt.Printf("  %-20s %-8s %-14s %-14s %-10s\n",
		"Тип", "sizeof", "HeapObjects", "Байт/объект", "Упаковка")
	fmt.Println("  ------------------------------------------------------------------")

	benchTiny := func(label string, sz uintptr, alloc func(int)) {
		runtime.GC()
		runtime.ReadMemStats(&before)
		alloc(N)
		runtime.ReadMemStats(&after)
		objD := after.HeapObjects - before.HeapObjects
		allocD := after.HeapAlloc - before.HeapAlloc
		pack := float64(N) / float64(objD)
		bpo := float64(allocD) / float64(N)
		fmt.Printf("  %-20s %-8d %-14d %-14.1f %-10.1fx\n",
			label, sz, objD, bpo, pack)
	}

	// bool = 1 байт
	var bools []*bool
	benchTiny("bool", unsafe.Sizeof(false), func(n int) {
		bools = make([]*bool, n)
		for i := range bools {
			v := true
			bools[i] = &v
		}
	})
	runtime.KeepAlive(bools)

	// int16 = 2 байта
	var i16s []*int16
	benchTiny("int16", unsafe.Sizeof(int16(0)), func(n int) {
		i16s = make([]*int16, n)
		for i := range i16s {
			v := int16(i)
			i16s[i] = &v
		}
	})
	runtime.KeepAlive(i16s)

	// int32 = 4 байта
	var i32s []*int32
	benchTiny("int32", unsafe.Sizeof(int32(0)), func(n int) {
		i32s = make([]*int32, n)
		for i := range i32s {
			v := int32(i)
			i32s[i] = &v
		}
	})
	runtime.KeepAlive(i32s)

	// int64 = 8 байт
	var i64s []*int64
	benchTiny("int64", unsafe.Sizeof(int64(0)), func(n int) {
		i64s = make([]*int64, n)
		for i := range i64s {
			v := int64(i)
			i64s[i] = &v
		}
	})
	runtime.KeepAlive(i64s)

	// [15]byte = 15 байт (ещё tiny)
	var b15s []*[15]byte
	benchTiny("[15]byte", unsafe.Sizeof([15]byte{}), func(n int) {
		b15s = make([]*[15]byte, n)
		for i := range b15s {
			b15s[i] = new([15]byte)
		}
	})
	runtime.KeepAlive(b15s)

	// [16]byte = 16 байт (уже НЕ tiny)
	var b16s []*[16]byte
	benchTiny("[16]byte", unsafe.Sizeof([16]byte{}), func(n int) {
		b16s = make([]*[16]byte, n)
		for i := range b16s {
			b16s[i] = new([16]byte)
		}
	})
	runtime.KeepAlive(b16s)

	// [24]byte = 24 байта
	var b24s []*[24]byte
	benchTiny("[24]byte", unsafe.Sizeof([24]byte{}), func(n int) {
		b24s = make([]*[24]byte, n)
		for i := range b24s {
			b24s[i] = new([24]byte)
		}
	})
	runtime.KeepAlive(b24s)

	fmt.Println()
	fmt.Println("  → Объекты < 16 байт без указателей: несколько штук в одном heap-объекте.")
	fmt.Println("  → Объекты >= 16 байт: ровно 1 heap-объект на каждый, упаковки нет.")
	fmt.Println("  → Чем меньше sizeof — тем больше упаковка (до 16/sizeof).")
}

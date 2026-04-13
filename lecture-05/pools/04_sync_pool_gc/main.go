// Пример 4: Объекты в sync.Pool умирают при GC.
//
// sync.Pool — это НЕ кэш. Рантайм может очистить пул в любой момент
// (исторически — каждый цикл GC, с Go 1.13 — через 2 цикла с victim cache).
//
// Это означает:
//   - Нельзя полагаться на то, что объект переживёт GC
//   - Pool оптимизирует аллокации, но не гарантирует переиспользование
//   - При длительном простое пул становится пустым
//
// Здесь мы наглядно показываем, как GC опустошает пул.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	fmt.Println("=== sync.Pool и сборщик мусора ===")
	fmt.Println()

	newCalls := 0
	pool := sync.Pool{
		New: func() any {
			newCalls++
			return fmt.Sprintf("объект #%d", newCalls)
		},
	}

	// Заполняем пул 5 объектами.
	fmt.Println("--- Заполняем пул ---")
	for range 5 {
		pool.Put(pool.Get()) // создаём через New и кладём обратно
	}
	fmt.Printf("  New() вызывался: %d раз\n", newCalls)

	// Достаём — объекты на месте.
	fmt.Println("\n--- Get без GC ---")
	for range 3 {
		obj := pool.Get()
		fmt.Printf("  Get: %v\n", obj)
		pool.Put(obj)
	}
	beforeGC := newCalls
	fmt.Printf("  New() вызовов: %d (не изменилось — объекты в пуле)\n", newCalls)

	// Принудительный GC — victim cache (Go 1.13+).
	// После первого GC объекты попадают в victim cache.
	// После второго — удаляются полностью.
	fmt.Println("\n--- Первый GC (→ victim cache) ---")
	runtime.GC()
	for range 3 {
		obj := pool.Get()
		fmt.Printf("  Get: %v\n", obj)
		pool.Put(obj)
	}
	fmt.Printf("  New() вызовов: %d (объекты ещё в victim cache)\n", newCalls)

	// Два GC подряд: после Get/Put выше объекты снова в primary pool.
	// 1-й GC: primary → victim (старый victim очищен).
	// 2-й GC: victim → удалён. Теперь пул полностью пуст.
	fmt.Println("\n--- Второй GC (victim cache → удалено) ---")
	runtime.GC()
	runtime.GC()

	fmt.Println("  Достаём 5 объектов:")
	for range 5 {
		obj := pool.Get()
		fmt.Printf("  Get: %v\n", obj)
	}
	fmt.Printf("  New() вызовов: %d (было %d → создано %d новых)\n",
		newCalls, beforeGC, newCalls-beforeGC)

	// Демонстрация: при непрерывном использовании объекты переиспользуются.
	fmt.Println("\n--- Непрерывное использование (без GC между Get/Put) ---")
	newCalls = 0
	pool2 := sync.Pool{
		New: func() any {
			newCalls++
			return make([]byte, 0, 1024)
		},
	}

	const rounds = 100
	for range rounds {
		buf := pool2.Get().([]byte)
		buf = buf[:0]
		pool2.Put(buf)
	}
	fmt.Printf("  %d Get/Put циклов, New() вызвался: %d раз\n", rounds, newCalls)
	fmt.Println("  (один раз создали — переиспользуем)")

	// То же самое, но с GC на каждом шаге.
	fmt.Println("\n--- Get/Put с GC на каждом шаге ---")
	newCalls = 0
	pool3 := sync.Pool{
		New: func() any {
			newCalls++
			return make([]byte, 0, 1024)
		},
	}

	for range 10 {
		buf := pool3.Get().([]byte)
		buf = buf[:0]
		pool3.Put(buf)
		runtime.GC()
		runtime.GC()
	}
	fmt.Printf("  10 Get/Put циклов с GC, New() вызвался: %d раз\n", newCalls)

	fmt.Println("\n  → Вывод: sync.Pool — оптимизация аллокаций, НЕ кэш.")
	fmt.Println("    Объекты могут быть собраны GC в любой момент простоя.")
}

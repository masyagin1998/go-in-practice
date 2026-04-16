package main

// Демонстрация влияния CGo-вызовов на планировщик Go.
//
// Ключевой факт: каждый CGo-вызов блокирует целый OS-тред.
// Go-планировщик не может прервать C-функцию (нет точек preemption).
//
// При GOMAXPROCS=1 у планировщика ровно один P (логический процессор).
// Пока C-функция работает на M (OS-тред), привязанном к этому P,
// другие горутины НЕ могут выполняться на нём.
//
// Однако Go runtime при CGo-вызове ОТВЯЗЫВАЕТ M от P:
// заблокированный тред уходит в C-код, а P передаётся новому M,
// чтобы другие горутины могли работать. Это называется «handoff».
//
// Создание нового OS-треда — не бесплатно, и при множестве
// параллельных CGo-вызовов система создаёт по одному потоку на каждый,
// независимо от GOMAXPROCS.
//
// В этом примере:
// - GOMAXPROCS=1 (один P)
// - N горутин одновременно вызывают busy_work (100ms каждая)
//
// Go-версия: CPU-bound цикл на одном P → горутины делят один процессор,
// суммарное время ≈ N * 100ms.
//
// CGo-версия: каждый вызов получает свой OS-тред (handoff),
// все N вызовов выполняются ПАРАЛЛЕЛЬНО. Время ≈ 100ms.

/*
#include "slow.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// goBusyWork — CPU-bound работа на ~ms миллисекунд.
// Использует цикл с вычислениями, а не wall-clock таймер,
// чтобы честно занимать CPU-время.
//
//go:noinline
func goBusyWork(ms int) {
	// Калибровка: подбираем число итераций под ~1ms.
	const itersPerMs = 800_000
	n := ms * itersPerMs
	x := 1.0
	for i := 0; i < n; i++ {
		x = x*1.00001 + 0.00001
	}
	// Используем x, чтобы компилятор не выкинул цикл.
	runtime.KeepAlive(x)
}

func main() {
	runtime.GOMAXPROCS(1)

	const (
		numGoroutines = 5
		workMs        = 100
	)

	// === Калибровка: замер одной goBusyWork ===
	t0 := time.Now()
	goBusyWork(workMs)
	singleGoTime := time.Since(t0)
	fmt.Printf("Калибровка: goBusyWork(%d) ≈ %v\n\n", workMs, singleGoTime.Round(time.Millisecond))

	// === Эксперимент 1: чистый Go (CPU-bound) ===
	// На одном P горутины выполняются по очереди.
	// Даже с асинхронным preemption (Go 1.14+) CPU-ресурс один —
	// суммарное CPU-время не может быть меньше N * workMs.
	fmt.Println("=== Чистый Go: CPU-bound на 1 P ===")
	start := time.Now()
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			t := time.Now()
			goBusyWork(workMs)
			fmt.Printf("  [Go] горутина %d: %v\n", id, time.Since(t).Round(time.Millisecond))
		}(i)
	}
	wg.Wait()
	goTotal := time.Since(start).Round(time.Millisecond)
	fmt.Printf("  Итого (Go):  %v (ожидаем ~%v: CPU-ресурс один)\n\n",
		goTotal, (singleGoTime * time.Duration(numGoroutines)).Round(time.Millisecond))

	// === Эксперимент 2: CGo-вызовы ===
	// Каждый CGo-вызов блокирует OS-тред. Runtime создаёт новые треды
	// через handoff, и все вызовы работают параллельно на разных ядрах.
	fmt.Println("=== CGo: busy_work на 1 P ===")
	start = time.Now()

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			t := time.Now()
			C.busy_work(C.int(workMs))
			fmt.Printf("  [CGo] горутина %d: %v\n", id, time.Since(t).Round(time.Millisecond))
		}(i)
	}
	wg.Wait()
	cgoTotal := time.Since(start).Round(time.Millisecond)
	fmt.Printf("  Итого (CGo): %v (ожидаем ~%dms: параллельно на %d OS-тредах)\n\n",
		cgoTotal, workMs, numGoroutines)

	// === Итог ===
	fmt.Println("=== Выводы ===")
	fmt.Printf("Go  (GOMAXPROCS=1, %d горутин): %v — последовательно, CPU-ресурс один\n",
		numGoroutines, goTotal)
	fmt.Printf("CGo (GOMAXPROCS=1, %d горутин): %v — параллельно на %d OS-тредах\n",
		numGoroutines, cgoTotal, numGoroutines)
	fmt.Println()
	fmt.Println("CGo-вызов блокирует OS-тред → runtime отвязывает M от P (handoff)")
	fmt.Println("и создаёт новый M для обслуживания P. Итог: GOMAXPROCS=1, но")
	fmt.Printf("реально работают %d OS-тредов параллельно.\n", numGoroutines)
	fmt.Println()
	fmt.Println("Это может быть неожиданным: GOMAXPROCS ограничивает Go-код,")
	fmt.Println("но НЕ ограничивает количество тредов в C-коде.")
}

package main

// Ограничение параллелизма CGo-вызовов через семафор на каналах.
//
// Проблема (показана в scheduler-cgo):
//   CGo-вызов блокирует OS-тред. При массовых CGo-вызовах runtime
//   создаёт по одному OS-треду на каждый вызов. Если тредов больше,
//   чем debug.SetMaxThreads (по умолчанию 10000) — fatal error.
//
// Решение:
//   Буферизированный канал struct{} как семафор — ограничивает число
//   одновременных CGo-вызовов.
//
// Паттерн:
//   sem <- struct{}{}   // захватить слот (блокируется, если буфер полон)
//   C.cpu_work(...)     // вызов C
//   <-sem               // освободить слот
//
// В этом примере ставим debug.SetMaxThreads(100) и запускаем 1000 задач.
// С семафором — работает. Без семафора — thread exhaustion.

/*
#include "work.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

const (
	numTasks    = 1000
	workMs      = 100
	threadLimit = 100
)

func main() {
	numCPU := runtime.NumCPU()

	fmt.Printf("CPU ядер:      %d\n", numCPU)
	fmt.Printf("Задач:         %d (по %dms каждая)\n", numTasks, workMs)
	fmt.Printf("Лимит тредов:  %d (debug.SetMaxThreads)\n\n", threadLimit)

	debug.SetMaxThreads(threadLimit)

	// === С семафором: не более NumCPU одновременных CGo-вызовов ===
	sem := make(chan struct{}, numCPU)

	fmt.Printf("=== С семафором (буфер = %d): макс %d одновременных CGo-вызовов ===\n",
		numCPU, numCPU)

	start := time.Now()
	var wg sync.WaitGroup

	for i := range numTasks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			C.cpu_work(C.int(workMs))
			<-sem
		}()
		_ = i
	}
	wg.Wait()
	semTime := time.Since(start).Round(time.Millisecond)

	batches := (numTasks + numCPU - 1) / numCPU
	expectedMs := batches * workMs
	fmt.Printf("  Время: %v (ожидаем ~%dms: %d батчей по %dms)\n", semTime, expectedMs, batches, workMs)
	fmt.Println("  ✓ Не упали — тредов хватило.")
	fmt.Println()

	// === Без семафора: thread exhaustion ===
	fmt.Printf("=== Без семафора: %d горутин, каждая блокирует OS-тред ===\n", numTasks)
	fmt.Printf("  Лимит тредов = %d — ожидаем crash...\n\n", threadLimit)

	for i := range numTasks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			C.cpu_work(C.int(workMs))
		}()
		_ = i
	}
	wg.Wait()

	// Сюда не дойдём — runtime.throw("thread exhaustion") — это fatal error,
	// recover его не ловит.
	fmt.Println("Если мы здесь — что-то пошло не так :)")
}

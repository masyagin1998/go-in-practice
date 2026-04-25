package main

// Фиксы к 04 и 05 через sync/atomic. atomic.Int64 для счётчика,
// atomic.Bool для флага. Обе операции вводят happens-before — компилятор
// не вынесет Load из цикла и не потеряет Store.

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func counterDemo() {
	var counter atomic.Int64
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1_000_000 {
				counter.Add(1)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("counter = %d (ожидалось 2000000)\n", counter.Load())
}

func flagDemo() {
	var done atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		spins := 0
		for !done.Load() {
			spins++
		}
		fmt.Printf("воркер вышел после %d спинов\n", spins)
	}()

	time.Sleep(100 * time.Millisecond)
	done.Store(true)
	wg.Wait()
}

func main() {
	fmt.Println("--- atomic.Int64 counter ---")
	counterDemo()
	fmt.Println("--- atomic.Bool flag ---")
	flagDemo()
}

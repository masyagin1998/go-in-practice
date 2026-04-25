package main

// Block-профиль и mutex-профиль: где ждём и где держим.
//
// 20 goroutine толкаются на одном sync.Mutex. Внутри критической
// секции — 1 ms "работы" (time.Sleep). Итого contention очевидный.
//
// Что в итоге в профилях:
//   block — сколько goroutine ждали Lock()
//   mutex — сколько goroutine держали Lock() (блокируя других)

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

func main() {
	// Включаем профилирование contention.
	// Block: семплируем каждое событие блокировки (>=1 нс).
	runtime.SetBlockProfileRate(1)
	// Mutex: fraction=N — каждое N-е событие.
	runtime.SetMutexProfileFraction(1)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				mu.Lock()
				time.Sleep(1 * time.Millisecond) // "работа" под мьютексом
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	dump := func(name string) {
		f, err := os.Create(name + ".pprof")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := pprof.Lookup(name).WriteTo(f, 0); err != nil {
			panic(err)
		}
	}
	dump("block")
	dump("mutex")

	fmt.Println("block.pprof и mutex.pprof сохранены")
}

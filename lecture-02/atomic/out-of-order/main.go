// go run main.go
//
// Демонстрация переупорядочивания инструкций (store-load reordering).
// Два потока: T1 пишет x=1, читает y; T2 пишет y=1, читает x.
// Если r1==0 && r2==0 — процессор (или компилятор) переупорядочил операции.
// На x86-64 это единственный допустимый вид реордеринга (store может быть
// отложен относительно последующего load).

package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(4)

	detected := 0

	for i := range 1_000_000 {
		var x, y int
		var r1, r2 int

		var wg sync.WaitGroup
		wg.Add(2)

		// Барьер для одновременного старта.
		var start sync.WaitGroup
		start.Add(1)

		go func() {
			runtime.LockOSThread()
			start.Wait()
			x = 1
			r1 = y
			wg.Done()
		}()

		go func() {
			runtime.LockOSThread()
			start.Wait()
			y = 1
			r2 = x
			wg.Done()
		}()

		start.Done()
		wg.Wait()

		if r1 == 0 && r2 == 0 {
			detected++
			fmt.Printf("⚠️  Out-of-order обнаружен! Итерация: %d\n", i)
			break
		}
	}

	if detected == 0 {
		fmt.Println("Out-of-order не обнаружен за 1 000 000 итераций.")
	}
}

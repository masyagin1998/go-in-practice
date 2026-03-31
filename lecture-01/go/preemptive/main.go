// Демонстрация: бесконечный for в горутине НЕ вешает программу.
//
// До Go 1.14 горутина с tight loop без вызовов функций не отдавала
// управление, и программа зависала. Начиная с Go 1.14 рантайм
// использует асинхронные сигналы (preemptive scheduling) для
// вытеснения таких горутин.
//
// GODEBUG=asyncpreemptoff=1 - отключить асинхронные сигналы
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	// Ограничиваем одним потоком, чтобы показать что даже на одном P
	// планировщик справляется.
	runtime.GOMAXPROCS(1)

	go func() {
		// Бесконечный tight loop — ни каналов, ни вызовов функций.
		for {
			_ = 1 + 1
		}
	}()

	// Если планировщик не вытеснит горутину выше, мы сюда не дойдём.
	for i := 1; i <= 5; i++ {
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("main: tick %d (goroutines: %d)\n", i, runtime.NumGoroutine())
	}

	fmt.Println("main: программа завершилась, не зависла!")
}

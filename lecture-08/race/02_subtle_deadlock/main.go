package main

// Неочевидный (частичный) deadlock: те же AB-BA на мьютексах, но в
// программе есть другая активность (тикер в main). Рантайм видит, что
// main runnable / просыпается по таймеру, → checkdead() молчит. -race
// тоже бесполезен: data race нет, есть просто два повисших Lock().
//
// Программа спокойно работает, "тикер тикает", две заклинившие goroutine
// — невидимые висяки. Ловится только goleak / goroutine pprof.

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var a, b sync.Mutex

	go func() {
		a.Lock()
		defer a.Unlock()
		time.Sleep(10 * time.Millisecond)
		b.Lock()
		defer b.Unlock()
	}()

	go func() {
		b.Lock()
		defer b.Unlock()
		time.Sleep(10 * time.Millisecond)
		a.Lock()
		defer a.Unlock()
	}()

	// Главная goroutine остаётся "живой" (просыпается по таймеру) →
	// рантайм не считает программу мёртвой. В реальной программе на
	// этом месте http-сервер, фоновый воркер, очередь задач.
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	for range 4 {
		<-tick.C
		fmt.Println("main жив, две goroutine тихо висят")
	}
	fmt.Println("выходим — две goroutine так и остались заблокированы")
}

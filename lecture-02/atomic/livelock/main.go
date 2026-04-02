// go run main.go
//
// Livelock: две горутины активны, но бесконечно уступают друг другу
// и не могут выполнить полезную работу.
// В отличие от deadlock, потоки НЕ заблокированы — они крутятся в цикле.
//
// Классический пример: два мьютекса, два потока берут их в разном порядке.
// Каждый берёт свой, пытается взять чужой (TryLock), не получается —
// отпускает свой и пробует снова. Как два человека в узком коридоре.

package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	mu1 sync.Mutex
	mu2 sync.Mutex

	yields1 atomic.Int64
	yields2 atomic.Int64
	work1   atomic.Int64
	work2   atomic.Int64
)

func goroutine1() {
	runtime.LockOSThread()
	for {
		mu1.Lock()
		// Держу mu1 подольше, чтобы горутина 2 успела взять mu2.
		time.Sleep(time.Microsecond)
		if !mu2.TryLock() {
			mu1.Unlock()
			yields1.Add(1)
			continue
		}
		work1.Add(1)
		mu2.Unlock()
		mu1.Unlock()
	}
}

func goroutine2() {
	runtime.LockOSThread()
	for {
		mu2.Lock()
		time.Sleep(time.Microsecond)
		if !mu1.TryLock() {
			mu2.Unlock()
			yields2.Add(1)
			continue
		}
		work2.Add(1)
		mu1.Unlock()
		mu2.Unlock()
	}
}

func main() {
	runtime.GOMAXPROCS(4)

	go goroutine1()
	go goroutine2()

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		y1, y2 := yields1.Load(), yields2.Load()
		w1, w2 := work1.Load(), work2.Load()
		fmt.Printf("[%ds] Г1: уступила %d, работала %d | Г2: уступила %d, работала %d\n",
			i+1, y1, w1, y2, w2)
	}

	y1, y2 := yields1.Load(), yields2.Load()
	w1, w2 := work1.Load(), work2.Load()
	total := y1 + y2 + w1 + w2
	fmt.Printf("\nИтого: %d уступок, %d полезных работ из %d итераций (%.1f%% впустую)\n",
		y1+y2, w1+w2, total, float64(y1+y2)/float64(total)*100)
}

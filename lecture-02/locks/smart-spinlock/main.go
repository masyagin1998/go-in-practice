// go run main.go
//
// «Умный» спинлок — ticket lock.
// Честный (fair) — порядок захвата соответствует порядку прихода.
// Вместо инструкции PAUSE (недоступна в Go) используем runtime.Gosched(),
// чтобы уступить процессор другим горутинам во время ожидания.

package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// TicketLock — честный спинлок на основе «билетиков».
type TicketLock struct {
	nextTicket atomic.Int32
	nowServing atomic.Int32
}

func (t *TicketLock) Lock() {
	ticket := t.nextTicket.Add(1) - 1
	for t.nowServing.Load() != ticket {
		runtime.Gosched() // аналог PAUSE — уступаем процессор
	}
}

func (t *TicketLock) Unlock() {
	t.nowServing.Add(1)
}

func main() {
	runtime.GOMAXPROCS(4)

	var lock TicketLock
	sharedCounter := 0

	var wg sync.WaitGroup

	for id := range 4 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runtime.LockOSThread()

			for range 100 {
				lock.Lock()
				fmt.Printf("Thread %d\n", id)
				sharedCounter++
				time.Sleep(10 * time.Millisecond)
				lock.Unlock()
			}
		}(id)
	}

	wg.Wait()
	fmt.Printf("Итоговое значение счётчика: %d (ожидается 400)\n", sharedCounter)
}

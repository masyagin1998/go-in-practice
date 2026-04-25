package main

// Очевидный deadlock: AB-BA на двух мьютексах + main ждёт wg. Без -race
// все goroutine паркуются → шедулер видит, что нет ни одной runnable, и
// падает с `fatal error: all goroutines are asleep - deadlock!`.
//
// Под -race детектор гонок держит фоновую служебную goroutine живой —
// `runtime.checkdead()` уже не срабатывает, и программа повисает молча.
// То есть `-race` не только не помогает с дедлоком, но и ломает встроенный
// детектор. Это известный артефакт TSan-рантайма (см. golang/go#13098).

import (
	"sync"
	"time"
)

func main() {
	var a, b sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		a.Lock()
		defer a.Unlock()
		time.Sleep(10 * time.Millisecond)
		b.Lock()
		defer b.Unlock()
	}()

	go func() {
		defer wg.Done()
		b.Lock()
		defer b.Unlock()
		time.Sleep(10 * time.Millisecond)
		a.Lock()
		defer a.Unlock()
	}()

	wg.Wait()
}

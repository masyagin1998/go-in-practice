package main

// Гонка на bool-флаге. По memory model Go это UB: без happens-before
// между писателем и читателем visibility не гарантирована, и компилятор
// формально имеет право закешировать `done` в регистре / вынести load
// из цикла. На практике на современном Go (1.14+ с async preemption)
// этого не происходит — каждая итерация содержит safepoint-проверку,
// которая мешает hoisting'у, и воркер обычно выходит.
//
// Главное: `-race` всё равно классифицирует это как DATA RACE. Код
// невалиден независимо от того, "повезло" ли вам сегодня — на другой
// архитектуре, другом компиляторе или после следующей оптимизации
// поведение поменяется молча.

import (
	"fmt"
	"sync"
	"time"
)

var done bool

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		spins := 0
		for !done {
			spins++
		}
		fmt.Printf("воркер вышел после %d спинов\n", spins)
	}()

	time.Sleep(100 * time.Millisecond)
	done = true

	// Подстраховка на случай, если компилятор всё-таки соптимизирует
	// load: не хотим зависнуть в демке навсегда.
	waitCh := make(chan struct{})
	go func() { wg.Wait(); close(waitCh) }()

	select {
	case <-waitCh:
		fmt.Println("успели — load из памяти не вырезан")
	case <-time.After(2 * time.Second):
		fmt.Println("воркер завис — load вырезан, спасает только -race")
	}
}

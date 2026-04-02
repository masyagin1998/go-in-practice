// go run main.go
//
// sync.WaitGroup — ожидание завершения группы горутин.
// Add(n) — сколько горутин ждём, Done() — горутина завершилась, Wait() — ждём всех.

package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Worker %d: начал\n", id)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Worker %d: закончил\n", id)
}

func main() {
	var wg sync.WaitGroup

	for i := range 5 {
		wg.Add(1)
		go worker(i, &wg)
	}

	wg.Wait()
	fmt.Println("Все воркеры завершились.")
}

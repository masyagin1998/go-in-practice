// go run main.go
//
// sync.Cond — условная переменная.
// Позволяет горутинам ждать наступления условия и быть разбуженными.
//
// Signal() — будит ОДНУ ждущую горутину.
// Broadcast() — будит ВСЕ ждущие горутины.

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	signalExample()
	fmt.Println()
	broadcastExample()
}

// --- Signal: producer будит одного consumer ---
func signalExample() {
	fmt.Println("=== Signal ===")

	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	queue := make([]int, 0, 10)

	// Consumer — ждёт элемент в очереди.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		for len(queue) == 0 {
			cond.Wait() // атомарно: Unlock + спим + Lock при пробуждении
		}
		item := queue[0]
		queue = queue[1:]
		mu.Unlock()
		fmt.Printf("Consumer: получил %d\n", item)
	}()

	// Producer — кладёт элемент и будит одного consumer.
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	queue = append(queue, 42)
	fmt.Println("Producer: положил 42")
	cond.Signal() // будим ОДНОГО
	mu.Unlock()

	wg.Wait()
}

// --- Broadcast: стартовый сигнал для всех ---
func broadcastExample() {
	fmt.Println("=== Broadcast ===")

	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	ready := false

	var wg sync.WaitGroup

	// 5 воркеров ждут сигнала "старт".
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mu.Lock()
			for !ready {
				cond.Wait()
			}
			mu.Unlock()
			fmt.Printf("Worker %d: стартовал!\n", id)
		}(i)
	}

	// Даём воркерам время встать в Wait.
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	ready = true
	fmt.Println("Main: даём старт всем!")
	cond.Broadcast() // будим ВСЕХ
	mu.Unlock()

	wg.Wait()
}

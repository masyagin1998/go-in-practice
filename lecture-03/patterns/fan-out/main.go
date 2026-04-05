// go run main.go
//
// Fan-Out: один канал с задачами раздаётся нескольким воркерам.
// Каждый воркер берёт следующую задачу из общего канала —
// задача достаётся ровно одному воркеру.

package main

import (
	"fmt"
	"math"
	"sync"
)

// генерируем числа для проверки
func generate(nums ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, n := range nums {
			ch <- n
		}
	}()
	return ch
}

// isPrime — наивная проверка простоты (имитация «тяжёлой» работы)
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	nums := generate(2, 3, 4, 17, 19, 20, 23, 97, 100, 101, 104729)

	// Fan-out: запускаем 3 воркера, каждый читает из того же канала
	results := make(chan string)
	var wg sync.WaitGroup

	for w := range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range nums {
				if isPrime(n) {
					results <- fmt.Sprintf("worker-%d: %d — простое", w, n)
				} else {
					results <- fmt.Sprintf("worker-%d: %d — составное", w, n)
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		fmt.Println(res)
	}
}

// go run main.go
//
// Worker pool: фиксированное число воркеров читают задачи из buffered канала.
// Буфер позволяет отправителю поставить пачку задач в очередь, не дожидаясь
// пока воркер освободится. Результаты собираются через отдельный канал.

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	const numWorkers = 3

	jobs := make(chan int, 5)    // очередь задач
	results := make(chan string) // результаты

	// Запускаем воркеров
	var wg sync.WaitGroup
	for w := range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Имитация работы разной длительности
				time.Sleep(time.Duration(50*(job%3+1)) * time.Millisecond)
				results <- fmt.Sprintf("worker-%d обработал задачу %d", w, job)
			}
		}()
	}

	// Закрываем results после завершения всех воркеров
	go func() {
		wg.Wait()
		close(results)
	}()

	// Отправляем задачи
	go func() {
		for i := range 10 {
			jobs <- i
		}
		close(jobs)
	}()

	// Собираем результаты
	for res := range results {
		fmt.Println(res)
	}
}

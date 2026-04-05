// go run main.go
//
// Timeout через select + time.After:
// если операция не завершилась за отведённое время — срабатывает таймаут.
// Показываем три сценария: быстрый ответ, медленный ответ, очень медленный.

package main

import (
	"fmt"
	"time"
)

// slowAPI имитирует вызов с переменной задержкой.
func slowAPI(delay time.Duration) <-chan string {
	ch := make(chan string, 1) // буфер 1, чтобы горутина не зависла при таймауте
	go func() {
		time.Sleep(delay)
		ch <- fmt.Sprintf("ответ (задержка %v)", delay)
	}()
	return ch
}

func main() {
	calls := []time.Duration{
		50 * time.Millisecond,  // быстрый — успеет
		150 * time.Millisecond, // средний — успеет впритык
		500 * time.Millisecond, // медленный — таймаут
	}

	for i, delay := range calls {
		fmt.Printf("--- вызов %d (задержка %v) ---\n", i, delay)

		select {
		case result := <-slowAPI(delay):
			fmt.Printf("  успех: %s\n", result)
		case <-time.After(200 * time.Millisecond):
			fmt.Printf("  таймаут! не дождались ответа за 200ms\n")
		}
	}
}

// go run main.go
//
// Context: управление временем жизни горутин через context.
// Три сценария:
// 1. context.WithTimeout — автоматическая отмена по таймауту
// 2. context.WithCancel  — ручная отмена
// 3. context.WithValue   — передача request-scoped данных

package main

import (
	"context"
	"fmt"
	"time"
)

// worker выполняет работу, пока контекст не отменён.
func worker(ctx context.Context, name string) {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("  [%s] остановлен: %v\n", name, ctx.Err())
			return
		default:
			fmt.Printf("  [%s] итерация %d\n", name, i)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// fetchData имитирует запрос, который уважает контекст.
func fetchData(ctx context.Context) (string, error) {
	select {
	case <-time.After(300 * time.Millisecond):
		return "данные получены", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	// 1. WithTimeout — отмена через 350ms
	fmt.Println("=== WithTimeout (350ms) ===")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel1()

	go worker(ctx1, "timeout-worker")
	time.Sleep(400 * time.Millisecond) // даём время увидеть отмену

	// 2. WithCancel — ручная отмена
	fmt.Println("\n=== WithCancel (отменяем через 250ms) ===")
	ctx2, cancel2 := context.WithCancel(context.Background())

	go worker(ctx2, "cancel-worker")
	time.Sleep(250 * time.Millisecond)
	cancel2() // явно отменяем
	time.Sleep(50 * time.Millisecond) // даём горутине напечатать сообщение об отмене

	// 3. WithValue — передача данных через контекст
	fmt.Println("\n=== WithValue ===")
	type requestIDKey struct{}
	ctx3 := context.WithValue(context.Background(), requestIDKey{}, "req-42")

	// Извлекаем значение из контекста
	if reqID, ok := ctx3.Value(requestIDKey{}).(string); ok {
		fmt.Printf("  request_id = %s\n", reqID)
	}

	// 4. Таймаут на конкретную операцию
	fmt.Println("\n=== Timeout на операцию (200ms на запрос, который идёт 300ms) ===")
	ctx4, cancel4 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel4()

	result, err := fetchData(ctx4)
	if err != nil {
		fmt.Printf("  ошибка: %v\n", err)
	} else {
		fmt.Printf("  результат: %s\n", result)
	}

	fmt.Println("\n=== Timeout на операцию (500ms на запрос, который идёт 300ms) ===")
	ctx5, cancel5 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel5()

	result, err = fetchData(ctx5)
	if err != nil {
		fmt.Printf("  ошибка: %v\n", err)
	} else {
		fmt.Printf("  результат: %s\n", result)
	}
}

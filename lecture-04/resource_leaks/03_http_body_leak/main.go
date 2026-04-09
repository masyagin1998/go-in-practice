// Утечка HTTP-соединений: если не закрыть resp.Body, TCP-соединение
// не будет возвращено в пул — новые запросы будут открывать новые соединения,
// пока не закончатся файловые дескрипторы или порты.
//
// Сборка:
//   go build -o 03_http_body_leak .
//
// Запуск:
//   ./03_http_body_leak
//
// Запуск с valgrind:
//   valgrind --tool=memcheck ./03_http_body_leak

package main

import (
	"fmt"
	"io"
	"net/http"
)

// fetchLeaky делает HTTP-запрос, но не закрывает тело ответа.
// Соединение не возвращается в пул transport'а.
func fetchLeaky(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// Забыли: defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
	// resp.Body не закрыт → TCP-соединение утекло.
}

// fetchCorrect — правильный вариант.
func fetchCorrect(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() // Обязательно закрываем тело!

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func main() {
	// Демонстрация: многократный запрос без закрытия Body.
	// В реальном приложении это приведёт к исчерпанию сокетов.
	url := "http://example.com"

	for i := 0; i < 5; i++ {
		result, err := fetchLeaky(url)
		if err != nil {
			fmt.Printf("Запрос %d: ошибка: %v\n", i, err)
			continue
		}
		fmt.Printf("Запрос %d: получено %d байт (Body не закрыт!)\n", i, len(result))
	}

	fmt.Println("\nПравильный вариант:")
	for i := 0; i < 5; i++ {
		result, err := fetchCorrect(url)
		if err != nil {
			fmt.Printf("Запрос %d: ошибка: %v\n", i, err)
			continue
		}
		fmt.Printf("Запрос %d: получено %d байт (Body закрыт)\n", i, len(result))
	}
}

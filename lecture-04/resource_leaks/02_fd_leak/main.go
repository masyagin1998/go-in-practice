// Утечка файловых дескрипторов: файлы открываются, но не закрываются.
// Каждый процесс имеет лимит на число открытых дескрипторов (ulimit -n).
// Когда лимит исчерпан, любой open/socket/accept вернёт ошибку.
//
// Сборка:
//   go build -o 02_fd_leak .
//
// Запуск:
//   ./02_fd_leak
//
// Запуск с valgrind:
//   valgrind --tool=memcheck ./02_fd_leak

package main

import (
	"fmt"
	"os"
)

// readFileLeaky открывает файл, читает его, но ЗАБЫВАЕТ закрыть.
// Файловый дескриптор утекает.
func readFileLeaky(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// Забыли: defer f.Close()

	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil {
		return nil, err // Дескриптор утёк — файл не закрыт!
	}
	return buf[:n], nil
	// Дескриптор утёк — файл не закрыт!
}

// readFileCorrect — правильный вариант с defer.
func readFileCorrect(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() // Гарантированно закроется при выходе из функции.

	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func main() {
	// Создаём временный файл для демонстрации.
	tmpFile, err := os.CreateTemp("", "leak-demo-*.txt")
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	tmpFile.WriteString("Тестовые данные")
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Открываем файл 10 000 раз без закрытия.
	// На системе с ulimit -n 1024 упадём примерно на ~1020-й итерации.
	for i := 0; i < 10000; i++ {
		_, err := readFileLeaky(tmpPath)
		if err != nil {
			fmt.Printf("Ошибка на итерации %d: %v\n", i, err)
			fmt.Println("Дескрипторы исчерпаны!")
			return
		}
	}

	fmt.Println("Все итерации завершены (лимит дескрипторов не достигнут).")
}

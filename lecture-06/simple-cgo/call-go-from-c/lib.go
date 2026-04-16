package main

// Экспортируем Go-функцию для вызова из C.
// Собирается как C-архив: go build -buildmode=c-archive.
// C-код получает заголовочный файл libgogreet.h с прототипами.
//
// C-точка входа (main) находится в cmd/main.c — отдельная директория,
// чтобы CGo не пытался компилировать .c файлы при сборке архива.

import "C"

import "fmt"

//export GoGreet
func GoGreet() {
	fmt.Println("GoGreet(): hello from Go!")
}

// main() обязателен для buildmode=c-archive, но не используется.
func main() {}

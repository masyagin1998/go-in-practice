// Пример 1: defer — отложенный вызов функции при выходе из scope.
//
// defer гарантирует, что функция будет вызвана при выходе из текущей функции,
// даже при раннем return или панике. Это Go-аналог RAII/cleanup из C++.
//
// Ключевые правила:
//   1. defer'ы выполняются в обратном порядке (LIFO — стек).
//   2. Аргументы defer вычисляются в момент вызова defer, а не в момент исполнения.
//   3. defer привязан к функции, а не к блоку {} (в отличие от деструкторов в C++).
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"os"
)

// demoOrder — defer'ы выполняются в обратном порядке (LIFO).
// Аналогия: стопка тарелок — последняя положенная снимается первой.
func demoOrder() {
	fmt.Println("=== Обратный порядок defer ===")

	for i := 1; i <= 5; i++ {
		defer fmt.Printf("  defer %d\n", i)
	}

	fmt.Println("  Конец функции.")
	// Вывод:
	//   Конец функции.
	//   defer 5
	//   defer 4
	//   defer 3
	//   defer 2
	//   defer 1
}

// demoFileClose — типичное использование: закрытие файла через defer.
// Гарантирует, что файл будет закрыт при любом пути выхода из функции.
func demoFileClose() {
	fmt.Println("\n=== defer для закрытия файла ===")

	// Создаём временный файл.
	f, err := os.CreateTemp("", "defer-demo-*.txt")
	if err != nil {
		fmt.Println("  Ошибка создания файла:", err)
		return
	}
	defer func() {
		name := f.Name()
		f.Close()
		os.Remove(name)
		fmt.Printf("  [defer] Файл %s закрыт и удалён.\n", name)
	}()

	// Пишем в файл.
	fmt.Fprintf(f, "Тестовые данные\n")
	fmt.Printf("  Записали в файл: %s\n", f.Name())

	// При выходе из функции defer закроет и удалит файл.
}

// demoScopeGotcha — defer привязан к ФУНКЦИИ, а не к блоку {}.
// Это частая ловушка: defer внутри вложенного {} не выполнится
// при выходе из блока — только при выходе из функции.
func demoScopeGotcha() {
	fmt.Println("\n=== defer и вложенные scope'ы ===")

	// Три вложенных scope'а, в каждом — своя переменная и свой defer.
	// Переменные объявлены через := и живут только внутри своего блока.
	// Но defer'ы выполнятся НЕ при выходе из блока, а при выходе из функции!
	{
		x := 10
		fmt.Printf("  Scope 1: x = %d\n", x)

		defer fmt.Printf("  [defer 1] x = %d (из scope 1)\n", x)

		{
			y := 20
			fmt.Printf("  Scope 2: y = %d\n", y)

			defer fmt.Printf("  [defer 2] y = %d (из scope 2)\n", y)

			{
				z := 30
				fmt.Printf("  Scope 3: z = %d\n", z)

				defer fmt.Printf("  [defer 3] z = %d (из scope 3)\n", z)

				// Здесь z выходит из scope — но defer НЕ выполнится сейчас!
			}
			// z уже недоступен, но defer 3 всё ещё ждёт.
			fmt.Println("  Вышли из scope 3 — defer 3 НЕ выполнился.")
		}
		// y уже недоступен, но defer 2 всё ещё ждёт.
		fmt.Println("  Вышли из scope 2 — defer 2 НЕ выполнился.")
	}
	// x уже недоступен, но defer 1 всё ещё ждёт.
	fmt.Println("  Вышли из scope 1 — defer 1 НЕ выполнился.")

	// defer fmt.Printf("  [defer 3] z = %d (из scope 3)\n", z)

	fmt.Println("  Конец функции — сейчас выполнятся все defer'ы:")
	// Все три defer'а выполнятся ЗДЕСЬ, в обратном порядке: 3, 2, 1.
}

func main() {
	demoOrder()
	demoFileClose()
	demoScopeGotcha()
}

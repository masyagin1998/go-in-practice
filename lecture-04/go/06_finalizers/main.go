// Пример 6: Финализаторы — runtime.SetFinalizer и runtime.AddCleanup.
//
// Финализатор — функция, которую GC вызывает перед сборкой объекта.
// Позволяет освободить внешние ресурсы (файлы, сокеты, C-память).
//
// Важно:
//   - Финализатор НЕ гарантирует, когда он будет вызван (зависит от GC).
//   - Финализатор НЕ гарантирует, что будет вызван вообще (при os.Exit — нет).
//   - SetFinalizer задерживает сборку объекта на один цикл GC.
//   - AddCleanup (с Go 1.24) — улучшенная замена: не задерживает сборку,
//     получает значение (не указатель), можно вешать несколько cleanup'ов.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
)

// === SetFinalizer (старый способ) ===

type Resource struct {
	Name string
	fd   int // Имитация файлового дескриптора.
}

func NewResource(name string, fd int) *Resource {
	r := &Resource{Name: name, fd: fd}

	// Регистрируем финализатор — GC вызовет его перед сборкой объекта.
	runtime.SetFinalizer(r, func(r *Resource) {
		fmt.Printf("  [SetFinalizer] Закрываю ресурс %q (fd=%d)\n", r.Name, r.fd)
		// Здесь был бы syscall.Close(r.fd) в реальном коде.
	})

	return r
}

// === AddCleanup (новый способ, Go 1.24+) ===

type Connection struct {
	Addr string
	id   int
}

func NewConnection(addr string, id int) *Connection {
	c := &Connection{Addr: addr, id: id}

	// AddCleanup не задерживает сборку объекта (в отличие от SetFinalizer).
	// Cleanup получает копию значения (int), а не указатель на объект —
	// это безопаснее и не мешает GC.
	runtime.AddCleanup(c, func(id int) {
		fmt.Printf("  [AddCleanup] Закрываю соединение id=%d\n", id)
	}, c.id)

	return c
}

func main() {
	fmt.Println("=== SetFinalizer ===")
	r := NewResource("database", 42)
	fmt.Printf("Создан ресурс: %s (fd=%d)\n", r.Name, r.fd)
	r = nil // Убираем ссылку — объект становится мусором.

	// GC запустит финализатор.
	runtime.GC()
	runtime.GC() // Второй цикл — SetFinalizer задерживает сборку на 1 цикл.

	fmt.Println("\n=== AddCleanup ===")
	c1 := NewConnection("localhost:5432", 1)
	c2 := NewConnection("localhost:6379", 2)
	fmt.Printf("Создано: %s (id=%d), %s (id=%d)\n", c1.Addr, c1.id, c2.Addr, c2.id)

	c1 = nil
	c2 = nil

	runtime.GC()

	fmt.Println("\n=== Предупреждение ===")
	fmt.Println("Финализаторы — это страховка, а не основной механизм очистки.")
	fmt.Println("Всегда используйте defer и явный Close() / Release().")
}

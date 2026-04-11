// XOR-связный список на Go с unsafe — GC собирает «скрытые» узлы.
//
// Go GC является точным (precise): он знает типы всех полей и отслеживает
// только настоящие указатели. Поле типа uintptr — это просто число,
// GC не считает его ссылкой. Поэтому узлы, доступные только через
// XOR-маскированные указатели, будут собраны сборщиком мусора.
//
// Мы доказываем это через runtime.AddCleanup (Go 1.24+): если GC вызвал
// cleanup-функцию узла — значит, он считает узел недостижимым и собрал его.

package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// Node — узел XOR-списка.
type Node struct {
	Value int
	Both  uintptr // prev XOR next — для GC это просто число, не указатель.
}

// collected считает количество собранных GC узлов.
var collected atomic.Int32

// XORList хранит голову и хвост списка.
type XORList struct {
	Head *Node
	Tail *Node
}

// PushBack добавляет элемент в конец списка.
func (l *XORList) PushBack(value int) {
	node := &Node{Value: value}

	// Регистрируем cleanup — он сработает, когда GC решит собрать узел.
	// В отличие от SetFinalizer, AddCleanup не воскрешает объект:
	// cleanup получает копию значения (int), а не указатель на объект.
	runtime.AddCleanup(node, func(val int) {
		fmt.Printf("    ⚠ Cleanup: узел %d собран GC!\n", val)
		collected.Add(1)
	}, node.Value)

	if l.Tail != nil {
		// node.both = tail XOR nil = tail
		node.Both = uintptr(unsafe.Pointer(l.Tail))
		// Обновляем both у старого tail: было (prev XOR nil), стало (prev XOR node).
		l.Tail.Both = l.Tail.Both ^ uintptr(unsafe.Pointer(node))
	} else {
		l.Head = node
	}
	l.Tail = node

	// ВАЖНО: после конвертации в uintptr, Go GC больше не видит ссылку
	// на node из других узлов. Единственные «настоящие» указатели —
	// l.Head и l.Tail. Промежуточные узлы удерживаются только через XOR.
}

// TraverseForward проходит список от головы к хвосту.
func (l *XORList) TraverseForward() []int {
	var result []int
	var prev *Node
	curr := l.Head

	for curr != nil && len(result) < 10 { // Лимит от бесконечного цикла.
		result = append(result, curr.Value)
		nextPtr := curr.Both ^ uintptr(unsafe.Pointer(prev))
		prev = curr
		curr = (*Node)(unsafe.Pointer(nextPtr))
	}
	return result
}

func main() {
	fmt.Println("=== XOR-список на Go + unsafe: GC собирает скрытые узлы ===")
	fmt.Println()

	list := &XORList{}

	// Заполняем список значениями 10, 20, 30, 40, 50.
	for i := 1; i <= 5; i++ {
		list.PushBack(i * 10)
	}
	fmt.Println("  Список создан: 10 → 20 → 30 → 40 → 50")
	fmt.Println()

	// До GC — всё работает (объекты ещё не собраны).
	fmt.Println("  [до GC]")
	fmt.Printf("  Проход вперёд: %v\n", list.TraverseForward())
	fmt.Println()

	// Принудительный вызов GC.
	// Промежуточные узлы (20, 30, 40) доступны только через uintptr XOR —
	// для GC это не ссылки, поэтому он считает эти узлы мусором.
	fmt.Println("  Вызываем runtime.GC()...")
	runtime.GC()
	// Даём cleanup-функциям отработать.
	runtime.Gosched()
	runtime.GC()
	runtime.Gosched()
	fmt.Println()

	n := collected.Load()
	fmt.Printf("  GC собрал %d узлов из 5.\n", n)
	fmt.Println()

	fmt.Println("  [после GC — проход 1]")
	fmt.Printf("  Проход вперёд: %v\n", list.TraverseForward())
	fmt.Println()

	fmt.Println("  [после GC — проход 2]")
	fmt.Printf("  Проход вперёд: %v\n", list.TraverseForward())
	fmt.Println()

	if n > 0 {
		fmt.Println("  ❌ GC считает промежуточные узлы недостижимыми!")
		fmt.Println("     Cleanup-функции доказывают: uintptr — не ссылка для GC.")
		fmt.Println("     В реальной программе обращение к собранным узлам — undefined behavior.")
	} else {
		fmt.Println("  Cleanup-функции не сработали (GC мог не успеть).")
		fmt.Println("  Но это НЕ безопасно — uintptr не удерживает объект.")
	}

	fmt.Println()
	fmt.Println("  Вывод: в Go нельзя прятать указатели в uintptr.")
	fmt.Println("  GC точный — он отслеживает только поля типа *T и unsafe.Pointer.")

	// Удерживаем list, чтобы Head и Tail не были собраны.
	runtime.KeepAlive(list)
}

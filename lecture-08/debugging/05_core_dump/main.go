package main

// Post-mortem отладка: снимаем gcore с живого процесса и анализируем
// в dlv core. Процесс продолжает работу после снятия дампа.

import (
	"fmt"
	"os"
	"time"
)

func process(name string, items []int) int {
	fmt.Printf("[process %q] старт, %d элементов\n", name, len(items))
	sum := 0
	for _, v := range items {
		sum += v
	}
	time.Sleep(60 * time.Second)
	return sum
}

func main() {
	fmt.Printf("[main] PID=%d; пока спим, запусти в другом терминале:\n",
		os.Getpid())
	fmt.Printf("  gcore %d\n", os.Getpid())
	fmt.Printf("  dlv core ./gobin core.%d\n", os.Getpid())

	result := process("batch-A", []int{1, 2, 3, 4, 5})
	fmt.Println("result:", result)
}

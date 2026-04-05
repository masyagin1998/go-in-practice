// go run main.go
//
// Tee channel: дублируем поток данных — каждый элемент из входного
// канала отправляется в оба выходных канала. Аналог команды tee в Unix.

package main

import "fmt"

// tee разветвляет входной канал на два выходных.
// Каждый элемент отправляется в оба канала.
func tee(in <-chan int) (<-chan int, <-chan int) {
	out1 := make(chan int)
	out2 := make(chan int)

	go func() {
		defer close(out1)
		defer close(out2)
		for v := range in {
			// Локальные копии для select — после отправки в один канал
			// обнуляем его, чтобы не отправить дважды туда же.
			o1, o2 := out1, out2
			for sent := 0; sent < 2; sent++ {
				select {
				case o1 <- v:
					o1 = nil
				case o2 <- v:
					o2 = nil
				}
			}
		}
	}()

	return out1, out2
}

func main() {
	// Источник данных
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range 5 {
			in <- i
		}
	}()

	// Разветвляем
	log, process := tee(in)

	// Читаем оба канала параллельно
	done := make(chan struct{})

	go func() {
		defer func() { done <- struct{}{} }()
		for v := range log {
			fmt.Printf("[log]     записали %d\n", v)
		}
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		for v := range process {
			fmt.Printf("[process] обработали %d (результат: %d)\n", v, v*v)
		}
	}()

	<-done
	<-done
}

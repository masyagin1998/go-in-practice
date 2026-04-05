// go run main.go
//
// Pipeline: цепочка стадий обработки, связанных каналами.
// Каждая стадия — горутина, которая читает из входного канала,
// трансформирует данные и пишет в выходной.

package main

import "fmt"

// generate создаёт канал с числами.
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// square возводит каждое число в квадрат.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// filter пропускает только числа больше порога.
func filter(in <-chan int, threshold int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n > threshold {
				out <- n
			}
		}
	}()
	return out
}

func main() {
	// Pipeline: generate → square → filter(>10)
	nums := generate(1, 2, 3, 4, 5, 6)
	squared := square(nums)
	big := filter(squared, 10)

	for v := range big {
		fmt.Println(v) // 16, 25, 36
	}
}

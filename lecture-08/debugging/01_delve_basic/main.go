package main

// Функция sumPositive содержит off-by-one: цикл стартует с i=1 и
// пропускает первый элемент. Задача — найти баг в dlv.

import "fmt"

func sumPositive(nums []int) int {
	sum := 0
	for i := 1; i < len(nums); i++ {
		if nums[i] > 0 {
			sum += nums[i]
		}
	}
	return sum
}

func main() {
	nums := []int{10, -3, 7, -1, 5}
	got := sumPositive(nums)
	want := 22 // 10 + 7 + 5
	fmt.Printf("sumPositive(%v) = %d, ожидалось %d\n", nums, got, want)
}

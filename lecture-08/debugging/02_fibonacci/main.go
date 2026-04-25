package main

// Три реализации числа Фибоначчи — удобный объект для отладки в delve:
// есть рекурсия (хорошо виден стек), мемоизация (map в locals) и итерация.

import "fmt"

func fibMemo(n int, cache map[int]int) int {
	if n <= 1 {
		return n
	}
	if v, ok := cache[n]; ok {
		return v
	}
	v := fibMemo(n-1, cache) + fibMemo(n-2, cache)
	cache[n] = v
	return v
}

func fibRec(n int) int {
	if n <= 1 {
		return n
	}
	return fibRec(n-1) + fibRec(n-2)
}

func fibIter(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func main() {
	const n = 5

	memo := fibMemo(n, make(map[int]int))
	rec := fibRec(n)
	iter := fibIter(n)

	fmt.Printf("fib(%d) memo = %d\n", n, memo)
	fmt.Printf("fib(%d) rec  = %d\n", n, rec)
	fmt.Printf("fib(%d) iter = %d\n", n, iter)
}

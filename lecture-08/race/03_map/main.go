package main

// У map есть встроенный детектор конкурентных записей. На каждом
// mapassign / mapdelete / mapaccess рантайм проверяет бит hashWriting в
// hmap.flags; при коллизии вызывает runtime.throw → процесс падает с
// `fatal error: concurrent map writes`. Работает без -race; не ловится
// recover'ом (это throw, а не panic).

import (
	"fmt"
	"sync"
)

func main() {
	m := map[int]int{}
	var wg sync.WaitGroup

	for i := range 4 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range 100_000 {
				m[id*1_000_000+j] = j
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("(сюда мы не дойдём)", len(m))
}

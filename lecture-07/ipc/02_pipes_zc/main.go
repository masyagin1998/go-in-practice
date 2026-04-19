package main

// 02_pipes_zc — то же, что 01_pipes, но peer пишет в pipe через
// vmsplice(SPLICE_F_GIFT): ядро не копирует байты из user-space в pipe-буфер,
// а забирает страницы пользователя "в дар". На стороне Go zero-copy
// невозможен без splice в другой fd — stdin всё равно копируется в Go heap.

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Printf("[go] получил %q\n", scanner.Text())
	}
}

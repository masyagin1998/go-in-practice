package main

// 01_pipes — простейший IPC: shell pipe `./peer | go run .`.
//
// Оболочка сама создаёт pipe(2), цепляет stdout C-процесса к stdin Go.
// Никакого fork+exec из самого Go.

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Printf("[go] принял: %s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
}

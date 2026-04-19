package main

// 04_mkfifo_zc — named pipe + vmsplice(SPLICE_F_GIFT) на стороне peer'а.
// То же, что 03_mkfifo (FIFO в ФС), но peer отправляет через vmsplice,
// Go читает обычным read. Демо одностороннее — Go только принимает.

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"
)

const fifoPath = "/tmp/ipc_fifo_zc"

func main() {
	_ = os.Remove(fifoPath)
	if err := syscall.Mkfifo(fifoPath, 0600); err != nil {
		log.Fatalf("mkfifo: %v", err)
	}
	defer os.Remove(fifoPath)

	fmt.Printf("FIFO %s создан. В другом терминале: ./peer\n", fifoPath)

	f, err := os.OpenFile(fifoPath, os.O_RDONLY, 0)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Printf("[go] получил %q\n", scanner.Text())
	}
}

package main

// 02_mkfifo — named pipe между Go и C. Не требует общего родителя:
// достаточно согласовать путь в ФС.
//
// open(2) на FIFO блокируется до появления партнёра — поэтому Go открывает
// оба конца параллельно в горутинах.

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"
)

const (
	fifoC2S    = "/tmp/ipc_fifo_c2s"
	fifoS2C    = "/tmp/ipc_fifo_s2c"
	iterations = 5
)

func main() {
	mustMkfifo(fifoC2S)
	mustMkfifo(fifoS2C)
	defer os.Remove(fifoC2S)
	defer os.Remove(fifoS2C)

	fmt.Println("FIFO созданы. В другом терминале: ./peer")
	fmt.Println("Ждём peer'а...")

	var wg sync.WaitGroup
	var writer, reader *os.File
	wg.Add(2)
	go func() { defer wg.Done(); writer = mustOpen(fifoC2S, os.O_WRONLY) }()
	go func() { defer wg.Done(); reader = mustOpen(fifoS2C, os.O_RDONLY) }()
	wg.Wait()
	defer writer.Close()
	defer reader.Close()

	buf := bufio.NewReader(reader)
	for i := range iterations {
		msg := fmt.Sprintf("ping server %d\n", i)
		fmt.Printf("[go] отправил %q\n", strings.TrimRight(msg, "\n"))
		if _, err := writer.WriteString(msg); err != nil {
			log.Fatalf("write: %v", err)
		}
		reply, err := buf.ReadString('\n')
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		fmt.Printf("[go] получил %q\n", strings.TrimRight(reply, "\n"))
	}
}

func mustMkfifo(path string) {
	_ = os.Remove(path)
	if err := syscall.Mkfifo(path, 0600); err != nil {
		log.Fatalf("mkfifo %s: %v", path, err)
	}
}

func mustOpen(path string, flag int) *os.File {
	f, err := os.OpenFile(path, flag, 0)
	if err != nil {
		log.Fatalf("open %s: %v", path, err)
	}
	return f
}

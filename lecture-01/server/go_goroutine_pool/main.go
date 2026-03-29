/*
Goroutine-pool TCP server (lock-free queue).
Usage: go_goroutine_pool <fib|sleep>
*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	numWorkers = 14
	lfqSize    = 4096
	lfqMask    = lfqSize - 1
)

var mode string

func fib(n uint64) uint64 {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return
	}
	line := strings.TrimSpace(scanner.Text())

	if mode == "fib" {
		n, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			fmt.Fprintln(conn, "error: expected integer N")
			return
		}
		fmt.Fprintln(conn, fib(n))
	} else {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(conn, 42)
	}
}

type connBox struct{ conn net.Conn }

var (
	lfqHead  atomic.Uint64
	lfqTail  atomic.Uint64
	lfqSlots [lfqSize]unsafe.Pointer
)

func lfqPush(conn net.Conn) {
	box := unsafe.Pointer(&connBox{conn})
	t := lfqTail.Load()
	for atomic.LoadPointer(&lfqSlots[t&lfqMask]) != nil {
		runtime.Gosched()
	}
	atomic.StorePointer(&lfqSlots[t&lfqMask], box)
	lfqTail.Store(t + 1)
}

func lfqPop() net.Conn {
	for {
		h := lfqHead.Load()
		p := atomic.LoadPointer(&lfqSlots[h&lfqMask])
		if p == nil {
			runtime.Gosched()
			continue
		}
		if lfqHead.CompareAndSwap(h, h+1) {
			atomic.StorePointer(&lfqSlots[h&lfqMask], nil)
			return (*connBox)(p).conn
		}
	}
}

func worker() {
	for {
		conn := lfqPop()
		handleClient(conn)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <fib|sleep>\n", os.Args[0])
		os.Exit(1)
	}
	mode = os.Args[1]
	runtime.GOMAXPROCS(14)

	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < numWorkers; i++ {
		go worker()
	}
	log.Printf("listening on :9001 (goroutine-pool %d workers, mode=%s)", numWorkers, mode)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		lfqPush(conn)
	}
}

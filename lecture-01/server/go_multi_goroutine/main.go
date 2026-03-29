/*
Multi-goroutine TCP server (goroutine-per-request).
Usage: go_multi_goroutine <fib|sleep>
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
	"time"
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
	log.Printf("listening on :9001 (multi-goroutine, mode=%s)", mode)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

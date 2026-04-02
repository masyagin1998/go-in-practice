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
	log.Printf("req from %s: %s", conn.RemoteAddr(), line)

	if mode == "fib" {
		n, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			log.Printf("resp to %s: error: expected integer N", conn.RemoteAddr())
			fmt.Fprintln(conn, "error: expected integer N")
			return
		}
		result := fib(n)
		log.Printf("resp to %s: %d", conn.RemoteAddr(), result)
		fmt.Fprintln(conn, result)
	} else {
		time.Sleep(100 * time.Millisecond)
		log.Printf("resp to %s: 42", conn.RemoteAddr())
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

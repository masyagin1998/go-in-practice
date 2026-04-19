package main

// 04_unix_sockets — AF_UNIX stream. Go — сервер, C — клиент.
// Как TCP, но не ходит через сетевой стек → latency ниже.

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	sockPath   = "/tmp/ipc_unix_socket"
	iterations = 5
)

func main() {
	_ = os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	defer os.Remove(sockPath)
	fmt.Printf("Слушаем %s. В другом терминале: ./peer\n", sockPath)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	defer conn.Close()

	rd := bufio.NewReader(conn)
	for range iterations {
		line, err := rd.ReadString('\n')
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		req := strings.TrimRight(line, "\n")
		fmt.Printf("[go] получил %q\n", req)

		reply := "echo: " + req
		fmt.Printf("[go] отправил %q\n", reply)
		if _, err := conn.Write([]byte(reply + "\n")); err != nil {
			log.Fatalf("write: %v", err)
		}
	}
}

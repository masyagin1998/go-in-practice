package main

// 03_sockets — TCP loopback. Go — сервер на 127.0.0.1:8891, C — клиент.
// Тот же API работает и между машинами.

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	addr       = "127.0.0.1:8891"
	iterations = 5
)

func main() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	fmt.Printf("Слушаем %s. В другом терминале: ./peer\n", addr)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	defer conn.Close()
	fmt.Printf("Подключился %s\n", conn.RemoteAddr())

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

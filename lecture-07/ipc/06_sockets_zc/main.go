package main

// 06_sockets_zc — TCP + MSG_ZEROCOPY на стороне peer'а. Go — сервер, читает
// обычным read; peer — клиент, шлёт через send(fd, ..., MSG_ZEROCOPY),
// чтобы ядро пинило его user-страницы и не копировало в skbuff.
//
// Демо симметрично 05_sockets: peer пишет "ping", Go отвечает "echo: ping".
// Реальный выигрыш ZC виден только на буферах ≥ ~16KB — для мелких ядро
// всё равно копирует (см. Documentation/networking/msg_zerocopy.rst).

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	addr       = "127.0.0.1:8892"
	iterations = 5
)

func main() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	fmt.Printf("Слушаем %s (MSG_ZEROCOPY на peer'е). В другом терминале: ./peer\n", addr)

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

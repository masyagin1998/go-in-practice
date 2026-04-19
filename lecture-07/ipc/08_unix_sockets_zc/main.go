package main

// 08_unix_sockets_zc — AF_UNIX + SCM_RIGHTS + memfd. Настоящий zero-copy:
// peer создаёт memfd (анонимный объект в памяти), пишет туда данные и
// передаёт сам fd через Unix-socket. Go получает fd, mmap'ит — оба
// процесса видят ОДНИ и те же физические страницы. Ни одной копии.
//
// В отличие от 10_shmem: нет named-объекта в /dev/shm, memfd живёт
// ровно пока открыт хоть у кого-то. Полностью ephemeral.

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
)

const (
	sockPath   = "/tmp/ipc_uds_zc"
	iterations = 5
)

func main() {
	_ = os.Remove(sockPath)
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: sockPath, Net: "unix"})
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	defer os.Remove(sockPath)

	fmt.Printf("Слушаем %s. В другом терминале: ./peer\n", sockPath)

	conn, err := ln.AcceptUnix()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	defer conn.Close()

	for range iterations {
		// Читаем 1-байтовый дамми + control-message со списком fd'ов.
		buf := make([]byte, 1)
		oob := make([]byte, syscall.CmsgSpace(4)) // один int
		_, oobn, _, _, err := conn.ReadMsgUnix(buf, oob)
		if err != nil {
			log.Fatalf("ReadMsgUnix: %v", err)
		}

		scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
		if err != nil {
			log.Fatalf("ParseSocketControlMessage: %v", err)
		}
		fds, err := syscall.ParseUnixRights(&scms[0])
		if err != nil {
			log.Fatalf("ParseUnixRights: %v", err)
		}

		fd := fds[0]
		// mmap'им fd → zero-copy view на те же физические страницы.
		var st syscall.Stat_t
		if err := syscall.Fstat(fd, &st); err != nil {
			log.Fatalf("fstat: %v", err)
		}
		page, err := syscall.Mmap(fd, 0, int(st.Size),
			syscall.PROT_READ, syscall.MAP_SHARED)
		if err != nil {
			log.Fatalf("mmap: %v", err)
		}
		fmt.Printf("[go] получил %q\n", string(page))
		_ = syscall.Munmap(page)
		_ = syscall.Close(fd)
	}
}

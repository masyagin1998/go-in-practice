package main

// 05_queue — POSIX message queue. Атомарные сообщения + приоритеты.
// Go создаёт очереди и инициирует обмен, C-peer слушает и эхо-отвечает.

/*
#cgo LDFLAGS: -lrt
#include <fcntl.h>
#include <mqueue.h>
#include <stdlib.h>

static mqd_t mq_create(const char *name, long maxmsg, long msgsize, int read_only) {
    struct mq_attr attr = {0};
    attr.mq_maxmsg  = maxmsg;
    attr.mq_msgsize = msgsize;
    int flags = (read_only ? O_RDONLY : O_WRONLY) | O_CREAT | O_EXCL;
    return mq_open(name, flags, 0600, &attr);
}

static int mq_is_err(mqd_t q) { return q == (mqd_t)-1; }
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

const (
	mqC2S      = "/ipc_mq_c2s"
	mqS2C      = "/ipc_mq_s2c"
	maxMsg     = 10
	msgSize    = 1024
	iterations = 5
)

func main() {
	cC2S := C.CString(mqC2S)
	cS2C := C.CString(mqS2C)
	defer C.free(unsafe.Pointer(cC2S))
	defer C.free(unsafe.Pointer(cS2C))

	C.mq_unlink(cC2S)
	C.mq_unlink(cS2C)

	c2s := C.mq_create(cC2S, C.long(maxMsg), C.long(msgSize), 0) // Go → peer
	s2c := C.mq_create(cS2C, C.long(maxMsg), C.long(msgSize), 1) // peer → Go
	if C.mq_is_err(c2s) != 0 || C.mq_is_err(s2c) != 0 {
		log.Fatal("mq_create failed")
	}
	defer C.mq_unlink(cC2S)
	defer C.mq_unlink(cS2C)
	defer C.mq_close(c2s)
	defer C.mq_close(s2c)

	fmt.Printf("Очереди созданы. В другом терминале: ./peer\n")

	for i := range iterations {
		msg := fmt.Sprintf("ping server %d", i)
		// Чётные i → prio=5, нечётные → prio=1. В пинг-понге 1×1 это не
		// меняет порядок доставки, но в реальной очереди prio=5 всегда
		// читается раньше prio=1.
		prio := C.uint(1)
		if i%2 == 0 {
			prio = 5
		}
		fmt.Printf("[go] отправил %q (prio=%d)\n", msg, prio)

		cmsg := C.CString(msg)
		rc := C.mq_send(c2s, cmsg, C.size_t(len(msg)), prio)
		C.free(unsafe.Pointer(cmsg))
		if rc != 0 {
			log.Fatalf("mq_send: rc=%d", rc)
		}

		buf := make([]byte, msgSize)
		var recvPrio C.uint
		n := C.mq_receive(s2c, (*C.char)(unsafe.Pointer(&buf[0])),
			C.size_t(len(buf)), &recvPrio)
		if n < 0 {
			log.Fatal("mq_receive")
		}
		fmt.Printf("[go] получил %q (prio=%d)\n", string(buf[:n]), recvPrio)
	}
}

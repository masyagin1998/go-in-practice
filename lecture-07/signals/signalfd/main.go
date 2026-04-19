package main

// signalfd — сигналы как файловый дескриптор. В отличие от signal.Notify,
// signalfd отдаёт полный struct signalfd_siginfo с payload'ом
// (ssi_int / ssi_ptr) и PID отправителя.
//
// Чтобы signalfd получал SIGUSR1/2 (а не Go-шный handler), сигналы
// должны быть заблокированы на ВСЕХ тредах. Блокируем в C-конструкторе —
// он выполняется до runtime·rt0_go, и все M-треды Go наследуют маску.

/*
#include <signal.h>
#include <sys/signalfd.h>

__attribute__((constructor))
static void block_usr(void) {
    sigset_t m;
    sigemptyset(&m);
    sigaddset(&m, SIGUSR1);
    sigaddset(&m, SIGUSR2);
    sigprocmask(SIG_BLOCK, &m, NULL);
}

static int make_signalfd(void) {
    sigset_t m;
    sigemptyset(&m);
    sigaddset(&m, SIGUSR1);
    sigaddset(&m, SIGUSR2);
    return signalfd(-1, &m, 0);
}
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"
)

const iterations = 5

func main() {
	fd := int(C.make_signalfd())
	if fd < 0 {
		log.Fatal("signalfd: -1")
	}
	defer syscall.Close(fd)

	fmt.Printf("PID=%d. В другом терминале: ./peer %d\n", os.Getpid(), os.Getpid())
	fmt.Println("Ждём 5 сигналов через signalfd...")

	siSize := int(unsafe.Sizeof(C.struct_signalfd_siginfo{}))
	buf := make([]byte, siSize)

	for range iterations {
		n, err := syscall.Read(fd, buf)
		if err != nil {
			log.Fatalf("read signalfd: %v", err)
		}
		if n != siSize {
			log.Fatalf("read signalfd: got %d, want %d", n, siSize)
		}
		si := (*C.struct_signalfd_siginfo)(unsafe.Pointer(&buf[0]))
		fmt.Printf("[go] signal=%s pid=%d ssi_int=%d\n",
			signame(uint32(si.ssi_signo)), int(si.ssi_pid), int(si.ssi_int))
	}
}

func signame(s uint32) string {
	switch s {
	case uint32(syscall.SIGUSR1):
		return "SIGUSR1"
	case uint32(syscall.SIGUSR2):
		return "SIGUSR2"
	default:
		return fmt.Sprintf("sig(%d)", s)
	}
}

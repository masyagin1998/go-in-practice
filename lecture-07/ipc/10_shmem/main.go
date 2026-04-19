package main

// 06_shmem — shared memory через shm_open + mmap.
// Go-клиент кладёт сообщение и поднимает флаг client_ready,
// C-сервер отвечает и поднимает server_ready. Busy-wait.
//
// Layout (1032 байта, см. c/shared.h):
//   [int32 client_ready][int32 server_ready][1024 байта buffer]
//
// cgo нужен только под shm_open/shm_unlink; памятью управляем через syscall.Mmap.

/*
#cgo LDFLAGS: -lrt
#include <fcntl.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

static int shm_create(const char *name, size_t size) {
    int fd = shm_open(name, O_CREAT | O_RDWR, 0600);
    if (fd < 0) return -1;
    if (ftruncate(fd, (off_t)size) < 0) { close(fd); shm_unlink(name); return -1; }
    return fd;
}
*/
import "C"

import (
	"fmt"
	"log"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	shmName    = "/ipc_shmem_demo"
	bufSize    = 1024
	shmSize    = 8 + bufSize
	iterations = 5
)

func main() {
	cName := C.CString(shmName)
	defer C.free(unsafe.Pointer(cName))

	C.shm_unlink(cName)
	fd := C.shm_create(cName, C.size_t(shmSize))
	if fd < 0 {
		log.Fatal("shm_create failed (проверьте /dev/shm и права)")
	}
	defer C.shm_unlink(cName)

	region, err := syscall.Mmap(int(fd), 0, shmSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap: %v", err)
	}
	defer syscall.Munmap(region)
	syscall.Close(int(fd))

	clientReady := (*atomic.Int32)(unsafe.Pointer(&region[0]))
	serverReady := (*atomic.Int32)(unsafe.Pointer(&region[4]))
	buffer := region[8 : 8+bufSize]

	fmt.Printf("Shmem %s создан. В другом терминале: ./peer\n", shmName)

	for i := range iterations {
		msg := fmt.Sprintf("ping server %d", i)
		copy(buffer, msg)
		buffer[len(msg)] = 0
		fmt.Printf("[go] отправил %q\n", msg)

		clientReady.Store(1)
		for serverReady.Load() == 0 {
			time.Sleep(100 * time.Microsecond)
		}
		fmt.Printf("[go] получил %q\n", cString(buffer))
		serverReady.Store(0)
	}
}

func cString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

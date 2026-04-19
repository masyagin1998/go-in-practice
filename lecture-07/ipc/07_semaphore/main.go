package main

// 07_semaphore — shmem + POSIX named semaphore.
// В отличие от 06_shmem busy-wait'а, ждём в ядре через sem_wait → 0% CPU.

/*
#cgo LDFLAGS: -lrt -lpthread
#include <fcntl.h>
#include <semaphore.h>
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

static sem_t *sem_create(const char *name) {
    sem_unlink(name);
    return sem_open(name, O_CREAT | O_EXCL, 0600, 0);
}

static int sem_wait_w(sem_t *s)  { return sem_wait(s); }
static int sem_post_w(sem_t *s)  { return sem_post(s); }
static int sem_close_w(sem_t *s) { return sem_close(s); }
*/
import "C"

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

const (
	shmName    = "/ipc_sem_demo"
	semClient  = "/ipc_sem_client"
	semServer  = "/ipc_sem_server"
	bufSize    = 1024
	iterations = 5
)

func main() {
	cShm := C.CString(shmName)
	cSc := C.CString(semClient)
	cSs := C.CString(semServer)
	defer C.free(unsafe.Pointer(cShm))
	defer C.free(unsafe.Pointer(cSc))
	defer C.free(unsafe.Pointer(cSs))

	C.shm_unlink(cShm)
	fd := C.shm_create(cShm, C.size_t(bufSize))
	if fd < 0 {
		log.Fatal("shm_create failed")
	}
	defer C.shm_unlink(cShm)

	sc := C.sem_create(cSc)
	ss := C.sem_create(cSs)
	if sc == nil || ss == nil {
		log.Fatal("sem_create failed")
	}
	defer C.sem_unlink(cSc)
	defer C.sem_unlink(cSs)
	defer C.sem_close_w(sc)
	defer C.sem_close_w(ss)

	region, err := syscall.Mmap(int(fd), 0, bufSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap: %v", err)
	}
	defer syscall.Munmap(region)
	syscall.Close(int(fd))

	fmt.Printf("Shmem+sems созданы. В другом терминале: ./peer\n")

	for i := range iterations {
		msg := fmt.Sprintf("ping server %d", i)
		copy(region, msg)
		region[len(msg)] = 0
		fmt.Printf("[go] отправляем %q\n", msg)

		C.sem_post_w(sc)  // "клиент готов"
		C.sem_wait_w(ss)  // ждём ответа сервера в ядре
		fmt.Printf("[go] ответ %q\n", cString(region))
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

package main

// mutex_shared — pthread_mutex с PTHREAD_PROCESS_SHARED в общей памяти.
// Go создаёт shmem + инициирует mutex, затем в 4 других терминалах
// запускаются C-воркеры. После Ctrl+C в Go — печатаем итоговый counter.

/*
#cgo LDFLAGS: -lrt -lpthread
#include <fcntl.h>
#include <pthread.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

typedef struct {
    pthread_mutex_t mutex;
    int counter;
} shared_data_t;

static int shm_create(const char *name, size_t size) {
    int fd = shm_open(name, O_CREAT | O_RDWR, 0600);
    if (fd < 0) return -1;
    if (ftruncate(fd, (off_t)size) < 0) { close(fd); shm_unlink(name); return -1; }
    return fd;
}

static int init_mutex(shared_data_t *d) {
    pthread_mutexattr_t attr;
    pthread_mutexattr_init(&attr);
    pthread_mutexattr_setpshared(&attr, PTHREAD_PROCESS_SHARED);
    int rc = pthread_mutex_init(&d->mutex, &attr);
    pthread_mutexattr_destroy(&attr);
    return rc;
}
*/
import "C"

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

const shmName = "/ipc_mutex_shared"

func main() {
	cName := C.CString(shmName)
	defer C.free(unsafe.Pointer(cName))

	C.shm_unlink(cName)
	size := C.size_t(unsafe.Sizeof(C.shared_data_t{}))
	fd := C.shm_create(cName, size)
	if fd < 0 {
		log.Fatal("shm_create")
	}
	defer C.shm_unlink(cName)

	region, err := syscall.Mmap(int(fd), 0, int(size),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap: %v", err)
	}
	defer syscall.Munmap(region)
	syscall.Close(int(fd))

	d := (*C.shared_data_t)(unsafe.Pointer(&region[0]))
	if rc := C.init_mutex(d); rc != 0 {
		log.Fatalf("pthread_mutex_init: rc=%d", rc)
	}
	d.counter = 0

	fmt.Println("Shmem + mutex готовы.")
	fmt.Println("Запустите C-воркеры в других терминалах:")
	fmt.Println("  ./peer 0   ./peer 1   ./peer 2   ./peer 3")
	fmt.Println("Каждый делает 100 инкрементов под mutex'ом.")
	fmt.Println("Ctrl+C для выхода с итоговым значением.")

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Printf("counter = %d\n", int(d.counter))
		case <-ctx.Done():
			fmt.Printf("\n[main] итоговый counter = %d\n", int(d.counter))
			return
		}
	}
}

package main

// mutex_death — восстановление после смерти владельца robust-mutex.
// C-peer берёт mutex и падает без unlock. Go пытается взять mutex,
// получает EOWNERDEAD и вызывает pthread_mutex_consistent.

/*
#cgo LDFLAGS: -lrt -lpthread
#include <errno.h>
#include <fcntl.h>
#include <pthread.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

typedef struct {
    pthread_mutex_t mutex;
    int value;
} shared_t;

static int shm_create(const char *name, size_t size) {
    int fd = shm_open(name, O_CREAT | O_RDWR, 0600);
    if (fd < 0) return -1;
    if (ftruncate(fd, (off_t)size) < 0) { close(fd); shm_unlink(name); return -1; }
    return fd;
}

static int init_robust(shared_t *s) {
    pthread_mutexattr_t a;
    pthread_mutexattr_init(&a);
    pthread_mutexattr_setpshared(&a, PTHREAD_PROCESS_SHARED);
    pthread_mutexattr_setrobust(&a, PTHREAD_MUTEX_ROBUST);
    int rc = pthread_mutex_init(&s->mutex, &a);
    pthread_mutexattr_destroy(&a);
    return rc;
}

// 0  — uncontended, peer mutex не держит
// 1  — EOWNERDEAD: peer умер, владение перешло к нам
// 2  — busy: peer держит mutex, попробуйте blocking lock
// <0 — -errno
static int trylock(shared_t *s) {
    int rc = pthread_mutex_trylock(&s->mutex);
    if (rc == 0)          return 0;
    if (rc == EOWNERDEAD) { pthread_mutex_consistent(&s->mutex); return 1; }
    if (rc == EBUSY)      return 2;
    return -rc;
}

static int lock(shared_t *s) {
    int rc = pthread_mutex_lock(&s->mutex);
    if (rc == EOWNERDEAD) { pthread_mutex_consistent(&s->mutex); return 1; }
    if (rc != 0)          return -rc;
    return 0;
}

static void unlock(shared_t *s) { pthread_mutex_unlock(&s->mutex); }
*/
import "C"

import (
	"fmt"
	"log"
	"syscall"
	"time"
	"unsafe"
)

const shmName = "/ipc_mutex_death"

func main() {
	cName := C.CString(shmName)
	defer C.free(unsafe.Pointer(cName))
	C.shm_unlink(cName)

	fd := C.shm_create(cName, C.size_t(unsafe.Sizeof(C.shared_t{})))
	if fd < 0 {
		log.Fatal("shm_create")
	}
	defer C.shm_unlink(cName)

	region, err := syscall.Mmap(int(fd), 0, int(unsafe.Sizeof(C.shared_t{})),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap: %v", err)
	}
	defer syscall.Munmap(region)
	syscall.Close(int(fd))

	s := (*C.shared_t)(unsafe.Pointer(&region[0]))
	if rc := C.init_robust(s); rc != 0 {
		log.Fatalf("init_robust: rc=%d", rc)
	}
	s.value = 0

	fmt.Println("Shmem + robust mutex готовы. В другом терминале: ./peer")
	fmt.Println("Peer возьмёт mutex и упадёт; мы опрашиваем trylock...")

	// Polling: ждём, пока peer не схватит mutex (trylock → EBUSY).
	// Затем blocking lock — он вернёт EOWNERDEAD, когда peer умрёт.
	for {
		switch rc := C.trylock(s); rc {
		case 0:
			// peer ещё не взял — отпускаем и ждём
			C.unlock(s)
			time.Sleep(50 * time.Millisecond)
		case 1:
			// EOWNERDEAD пойман на trylock'е (peer успел и взять, и упасть)
			handleRecovered(s)
			return
		case 2:
			// peer держит mutex — блокирующий lock дождётся его смерти
			if rc2 := C.lock(s); rc2 == 1 {
				handleRecovered(s)
			} else {
				fmt.Printf("[go] lock вернул %d (ожидали EOWNERDEAD)\n", rc2)
				C.unlock(s)
			}
			return
		default:
			log.Fatalf("[go] trylock error: errno=%d", -int(rc))
		}
	}
}

func handleRecovered(s *C.shared_t) {
	fmt.Printf("[go] поймали EOWNERDEAD, value=%d → восстанавливаем\n", int(s.value))
	s.value = 99
	C.unlock(s)
	fmt.Printf("[go] после восстановления value=%d\n", int(s.value))
}

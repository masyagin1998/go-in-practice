package main

// thread-masking — показать, что сигнал ловит только тред с
// разблокированным сигналом.
//
// Модель: kill(pid, SIGUSR1) — process-directed signal. Ядро выберет
// ЛЮБОЙ тред процесса, у которого SIGUSR1 не заблокирован. Если
// неблокированных тредов несколько — распределение произвольное (на
// Linux обычно round-robin или first-available).
//
// Демо: C-конструктор блокирует SIGUSR1 на главном треде → все Go M-треды
// наследуют маску с блоком. Затем 4 goroutine'ы пинятся через LockOSThread,
// первые две ОСТАВЛЯЮТ блок, последние две — pthread_sigmask UNBLOCK.
// После N сигналов видно, что счётчики растут только у 2 и 3.

/*
#include <signal.h>
#include <pthread.h>

#define N_THREADS 4

static pthread_t thread_ids[N_THREADS];
static volatile long catches[N_THREADS] = {0};

// Блокируем SIGUSR1 на главном треде ДО старта Go-runtime'а; все
// последующие M-треды унаследуют маску.
__attribute__((constructor))
static void block_sigusr1(void) {
    sigset_t m;
    sigemptyset(&m);
    sigaddset(&m, SIGUSR1);
    pthread_sigmask(SIG_BLOCK, &m, NULL);
}

static int which_thread(pthread_t me) {
    for (int i = 0; i < N_THREADS; i++) {
        if (pthread_equal(thread_ids[i], me)) return i;
    }
    return -1;
}

static void handler(int sig) {
    (void)sig;
    int i = which_thread(pthread_self());
    if (i >= 0) __sync_fetch_and_add(&catches[i], 1);
}

static void install_handler(void) {
    struct sigaction sa = {0};
    sa.sa_handler = handler;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_RESTART;
    sigaction(SIGUSR1, &sa, NULL);
}

static void register_thread(int i) { thread_ids[i] = pthread_self(); }

static void unblock_sigusr1(void) {
    sigset_t m;
    sigemptyset(&m);
    sigaddset(&m, SIGUSR1);
    pthread_sigmask(SIG_UNBLOCK, &m, NULL);
}

static long get_catches(int i) { return catches[i]; }
*/
import "C"

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

const threads = 4

func main() {
	C.install_handler()

	// Стартуем 4 goroutine'ы на отдельных OS-тредах; первые 2 оставляют
	// блок SIGUSR1, последние 2 — снимают.
	ready := sync.WaitGroup{}
	ready.Add(threads)
	for i := range threads {
		go func(id int) {
			runtime.LockOSThread() // пинуемся, чтобы pthread_self() был стабилен
			C.register_thread(C.int(id))
			if id >= 2 {
				C.unblock_sigusr1()
			}
			ready.Done()
			// Сидим вечно — тред должен быть живой, когда прилетит сигнал.
			select {}
		}(i)
	}
	ready.Wait()

	pid := os.Getpid()
	fmt.Printf("PID=%d, тредов-претендентов: %d\n", pid, threads)
	fmt.Printf("  thread 0: SIGUSR1 заблокирован\n")
	fmt.Printf("  thread 1: SIGUSR1 заблокирован\n")
	fmt.Printf("  thread 2: SIGUSR1 разблокирован\n")
	fmt.Printf("  thread 3: SIGUSR1 разблокирован\n")
	fmt.Printf("\nВ другом терминале: ./peer %d 100\n", pid)
	fmt.Printf("После этого Ctrl+C — напечатаем распределение.\n\n")

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	fmt.Println("\nРезультат:")
	for i := range threads {
		blocked := "разблокирован"
		if i < 2 {
			blocked = "заблокирован  "
		}
		fmt.Printf("  thread %d (%s): поймал %d сигналов\n",
			i, blocked, int64(C.get_catches(C.int(i))))
	}
	fmt.Println("\nОжидаем: 0+0 у заблокированных, сумма у разблокированных ≈ 100.")
}

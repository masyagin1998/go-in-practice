/* Thread-pool TCP server (lock-free queue). Usage: ./c_thread_pool <fib|sleep> */

#include "../c_lib/server_lib.h"
#include <pthread.h>
#include <stdatomic.h>

#define NUM_WORKERS 14
#define LFQ_SIZE 4096
#define LFQ_MASK (LFQ_SIZE - 1)

struct client_args { int fd; struct sockaddr_in addr; };

static _Alignas(64) atomic_uint_fast64_t lfq_head;
static _Alignas(64) atomic_uint_fast64_t lfq_tail;
static _Atomic(struct client_args *) lfq_slots[LFQ_SIZE];

static void lfq_push(struct client_args *ca) {
    uint_fast64_t t = atomic_load_explicit(&lfq_tail, memory_order_relaxed);
    while (atomic_load_explicit(&lfq_slots[t & LFQ_MASK], memory_order_acquire) != NULL) ;
    atomic_store_explicit(&lfq_slots[t & LFQ_MASK], ca, memory_order_release);
    atomic_store_explicit(&lfq_tail, t + 1, memory_order_release);
}

static struct client_args *lfq_pop(void) {
    for (;;) {
        uint_fast64_t h = atomic_load_explicit(&lfq_head, memory_order_relaxed);
        struct client_args *ca = atomic_load_explicit(&lfq_slots[h & LFQ_MASK], memory_order_acquire);
        if (ca == NULL) continue;
        if (atomic_compare_exchange_weak_explicit(&lfq_head, &h, h + 1,
                memory_order_acq_rel, memory_order_relaxed)) {
            atomic_store_explicit(&lfq_slots[h & LFQ_MASK], NULL, memory_order_release);
            return ca;
        }
    }
}

static void *worker_thread(void *arg) {
    (void)arg;
    for (;;) {
        struct client_args *ca = lfq_pop();
        handle_client(ca->fd, &ca->addr);
        free(ca);
    }
}

int main(int argc, char **argv) {
    parse_mode(argc, argv);
    int srv = server_listen();

    for (int i = 0; i < NUM_WORKERS; i++) {
        pthread_t tid;
        pthread_create(&tid, NULL, worker_thread, NULL);
        pthread_detach(tid);
    }

    log_msg("listening on :%d (thread-pool %d workers, mode=%s)", PORT, NUM_WORKERS, argv[1]);

    for (;;) {
        struct client_args *ca = malloc(sizeof(*ca));
        if (!ca) continue;
        socklen_t len = sizeof(ca->addr);
        ca->fd = accept(srv, (struct sockaddr *)&ca->addr, &len);
        if (ca->fd < 0) { free(ca); continue; }
        lfq_push(ca);
    }
}

#define _GNU_SOURCE

/*
 * Thread-pool + epoll TCP server (full async I/O).
 * Usage: ./c_thread_pool_epoll <fib|sleep>
 *
 * epoll thread handles ALL I/O. Workers ONLY compute (fib or nanosleep).
 * Workers signal back via eventfd.
 */

#include "../c_lib/server_lib.h"
#include <fcntl.h>
#include <pthread.h>
#include <stdatomic.h>
#include <sys/epoll.h>
#include <sys/eventfd.h>

#define NUM_WORKERS 14
#define LFQ_SIZE 4096
#define LFQ_MASK (LFQ_SIZE - 1)
#define MAX_EVENTS 256

/* work: epoll -> workers */
struct work_item { int fd; uint64_t val; };

/* done: workers -> epoll */
struct done_item { int fd; char resp[32]; int resp_len; };

/* SPMC work queue */
static _Alignas(64) atomic_uint_fast64_t wh, wt;
static _Atomic(struct work_item *) ws[LFQ_SIZE];

static void work_push(struct work_item *w) {
    uint_fast64_t t = atomic_load_explicit(&wt, memory_order_relaxed);
    while (atomic_load_explicit(&ws[t & LFQ_MASK], memory_order_acquire) != NULL) ;
    atomic_store_explicit(&ws[t & LFQ_MASK], w, memory_order_release);
    atomic_store_explicit(&wt, t + 1, memory_order_release);
}

static struct work_item *work_pop(void) {
    for (;;) {
        uint_fast64_t h = atomic_load_explicit(&wh, memory_order_relaxed);
        struct work_item *w = atomic_load_explicit(&ws[h & LFQ_MASK], memory_order_acquire);
        if (!w) continue;
        if (atomic_compare_exchange_weak_explicit(&wh, &h, h + 1,
                memory_order_acq_rel, memory_order_relaxed)) {
            atomic_store_explicit(&ws[h & LFQ_MASK], NULL, memory_order_release);
            return w;
        }
    }
}

/* MPSC done queue */
static _Alignas(64) atomic_uint_fast64_t dh, dt;
static _Atomic(struct done_item *) ds[LFQ_SIZE];

static void done_push(struct done_item *d) {
    uint_fast64_t t;
    for (;;) {
        t = atomic_load_explicit(&dt, memory_order_relaxed);
        if (atomic_compare_exchange_weak_explicit(&dt, &t, t + 1,
                memory_order_acq_rel, memory_order_relaxed)) break;
    }
    while (atomic_load_explicit(&ds[t & LFQ_MASK], memory_order_acquire) != NULL) ;
    atomic_store_explicit(&ds[t & LFQ_MASK], d, memory_order_release);
}

static struct done_item *done_try_pop(void) {
    uint_fast64_t h = atomic_load_explicit(&dh, memory_order_relaxed);
    struct done_item *d = atomic_load_explicit(&ds[h & LFQ_MASK], memory_order_acquire);
    if (!d) return NULL;
    atomic_store_explicit(&ds[h & LFQ_MASK], NULL, memory_order_release);
    atomic_store_explicit(&dh, h + 1, memory_order_release);
    return d;
}

static int evfd;

static void *worker_thread(void *arg) {
    (void)arg;
    for (;;) {
        struct work_item *w = work_pop();
        struct done_item *d = malloc(sizeof(*d));
        d->fd = w->fd;

        if (g_mode == MODE_FIB) {
            uint64_t result = fib(w->val);
            d->resp_len = snprintf(d->resp, sizeof(d->resp), "%" PRIu64 "\n", result);
        } else {
            sleep_ms();
            d->resp_len = snprintf(d->resp, sizeof(d->resp), "42\n");
        }
        free(w);
        done_push(d);
        uint64_t one = 1;
        write(evfd, &one, sizeof(one));
    }
}

enum ev_type { EV_LISTENER, EV_CLIENT, EV_DONE };
struct ev_data { enum ev_type type; int fd; };

static void set_nonblocking(int fd) {
    fcntl(fd, F_SETFL, fcntl(fd, F_GETFL, 0) | O_NONBLOCK);
}

int main(int argc, char **argv) {
    parse_mode(argc, argv);
    int srv = server_listen();
    set_nonblocking(srv);

    evfd = eventfd(0, EFD_NONBLOCK);
    int epfd = epoll_create1(0);

    struct ev_data ev_l = { .type = EV_LISTENER, .fd = srv };
    epoll_ctl(epfd, EPOLL_CTL_ADD, srv, &(struct epoll_event){ .events = EPOLLIN, .data.ptr = &ev_l });

    struct ev_data ev_d = { .type = EV_DONE, .fd = evfd };
    epoll_ctl(epfd, EPOLL_CTL_ADD, evfd, &(struct epoll_event){ .events = EPOLLIN, .data.ptr = &ev_d });

    for (int i = 0; i < NUM_WORKERS; i++) {
        pthread_t tid;
        pthread_create(&tid, NULL, worker_thread, NULL);
        pthread_detach(tid);
    }

    log_msg("listening on :%d (thread-pool %d + epoll + eventfd, mode=%s)", PORT, NUM_WORKERS, argv[1]);

    struct epoll_event events[MAX_EVENTS];
    for (;;) {
        int nfds = epoll_wait(epfd, events, MAX_EVENTS, -1);
        if (nfds < 0) { if (errno == EINTR) continue; break; }

        for (int i = 0; i < nfds; i++) {
            struct ev_data *ed = events[i].data.ptr;

            if (ed->type == EV_LISTENER) {
                for (;;) {
                    struct ev_data *cl = malloc(sizeof(*cl));
                    if (!cl) break;
                    cl->type = EV_CLIENT;
                    struct sockaddr_in addr;
                    socklen_t len = sizeof(addr);
                    cl->fd = accept(srv, (struct sockaddr *)&addr, &len);
                    if (cl->fd < 0) { free(cl); break; }
                    set_nonblocking(cl->fd);
                    epoll_ctl(epfd, EPOLL_CTL_ADD, cl->fd,
                              &(struct epoll_event){ .events = EPOLLIN, .data.ptr = cl });
                }
            } else if (ed->type == EV_CLIENT) {
                char buf[BUF_SIZE];
                ssize_t n = read(ed->fd, buf, sizeof(buf) - 1);
                if (n <= 0) {
                    epoll_ctl(epfd, EPOLL_CTL_DEL, ed->fd, NULL);
                    close(ed->fd); free(ed); continue;
                }
                buf[n] = '\0';
                log_msg("req fd=%d: %s", ed->fd, buf);
                epoll_ctl(epfd, EPOLL_CTL_DEL, ed->fd, NULL);

                struct work_item *w = malloc(sizeof(*w));
                w->fd = ed->fd;
                w->val = 0;
                if (g_mode == MODE_FIB) sscanf(buf, "%" SCNu64, &w->val);
                work_push(w);
                free(ed);
            } else {
                uint64_t count;
                read(evfd, &count, sizeof(count));
                struct done_item *d;
                while ((d = done_try_pop()) != NULL) {
                    log_msg("resp fd=%d: %.*s", d->fd, d->resp_len - 1, d->resp);
                    write(d->fd, d->resp, d->resp_len);
                    close(d->fd);
                    free(d);
                }
            }
        }
    }
}

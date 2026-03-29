#define _GNU_SOURCE

/*
 * Coroutine-based TCP server (single-thread, epoll + timerfd + green threads).
 * Usage: ./c_coroutine <fib|sleep>
 *
 * This is essentially a hand-written Go runtime:
 *   - Coroutines = goroutines (cooperative green threads)
 *   - Scheduler  = Go's M:N scheduler (single-threaded here)
 *   - coro_read/coro_write = Go's net.Conn.Read/Write (yield on EAGAIN)
 *   - coro_sleep = time.Sleep (timerfd + yield)
 *
 * handle_client() reads like plain sequential code — no callbacks,
 * no state machines, no async/await. Just like Go.
 *
 * gcc -std=gnu11 -O2 -Wall -Wextra -o c_coroutine c_coroutine.c ../c_lib/server_lib.c
 */

#include "../c_lib/server_lib.h"
#include <fcntl.h>
#include <sys/epoll.h>
#include <sys/timerfd.h>

/* ================================================================
 * Coroutine runtime
 * ================================================================ */

#define MAX_COROS  4096
#define CORO_STACK (64 * 1024)
#define MAX_EVENTS 256

typedef struct Coro {
    void     *rsp;
    uint64_t  regs[6]; /* rbx, rbp, r12, r13, r14, r15 */
    void     *stack_base;
    int       client_fd;
    int       active;
} Coro;

/* x86-64 SysV ABI: rdi = from, rsi = to */
__attribute__((naked))
static void ctx_switch(Coro *from, Coro *to) {
    (void)from; (void)to;
    asm(
        "movq %rbx,  8(%rdi)\n\t"
        "movq %rbp, 16(%rdi)\n\t"
        "movq %r12, 24(%rdi)\n\t"
        "movq %r13, 32(%rdi)\n\t"
        "movq %r14, 40(%rdi)\n\t"
        "movq %r15, 48(%rdi)\n\t"
        "movq %rsp,  0(%rdi)\n\t"

        "movq  8(%rsi), %rbx\n\t"
        "movq 16(%rsi), %rbp\n\t"
        "movq 24(%rsi), %r12\n\t"
        "movq 32(%rsi), %r13\n\t"
        "movq 40(%rsi), %r14\n\t"
        "movq 48(%rsi), %r15\n\t"
        "movq  0(%rsi), %rsp\n\t"

        "ret\n\t"
    );
}

static Coro  sched_ctx;             /* scheduler context               */
static Coro *current;               /* currently running coroutine     */
static Coro  pool[MAX_COROS];       /* coroutine pool                  */
static int   epfd;                  /* epoll fd                        */

/* free list (stack of indices) */
static int free_stack[MAX_COROS];
static int free_top;

/* ready queue (ring buffer) */
static Coro *ready_q[MAX_COROS];
static int rq_h, rq_t, rq_n;

static void ready_push(Coro *c) {
    ready_q[rq_t] = c;
    rq_t = (rq_t + 1) % MAX_COROS;
    rq_n++;
}

static Coro *ready_pop(void) {
    Coro *c = ready_q[rq_h];
    rq_h = (rq_h + 1) % MAX_COROS;
    rq_n--;
    return c;
}

static void coro_yield(void) {
    ctx_switch(current, &sched_ctx);
}

/* ================================================================
 * Coroutine I/O — looks blocking, but yields to scheduler
 * ================================================================ */

static void set_nonblocking(int fd) {
    fcntl(fd, F_SETFL, fcntl(fd, F_GETFL, 0) | O_NONBLOCK);
}

static ssize_t coro_read(int fd, void *buf, size_t len) {
    for (;;) {
        ssize_t n = read(fd, buf, len);
        if (n >= 0) return n;
        if (errno != EAGAIN && errno != EWOULDBLOCK) return n;
        struct epoll_event ev = { .events = EPOLLIN | EPOLLONESHOT,
                                  .data.ptr = current };
        epoll_ctl(epfd, EPOLL_CTL_ADD, fd, &ev);
        coro_yield();
        epoll_ctl(epfd, EPOLL_CTL_DEL, fd, NULL);
    }
}

static ssize_t coro_write(int fd, const void *buf, size_t len) {
    size_t done = 0;
    while (done < len) {
        ssize_t n = write(fd, (const char *)buf + done, len - done);
        if (n > 0) { done += n; continue; }
        if (n == 0) return done;
        if (errno != EAGAIN && errno != EWOULDBLOCK) return -1;
        struct epoll_event ev = { .events = EPOLLOUT | EPOLLONESHOT,
                                  .data.ptr = current };
        epoll_ctl(epfd, EPOLL_CTL_ADD, fd, &ev);
        coro_yield();
        epoll_ctl(epfd, EPOLL_CTL_DEL, fd, NULL);
    }
    return done;
}

static void coro_sleep(int ms) {
    int tfd = timerfd_create(CLOCK_MONOTONIC, TFD_NONBLOCK);
    struct itimerspec ts = {
        .it_value = { .tv_sec = ms / 1000,
                      .tv_nsec = (ms % 1000) * 1000000L }
    };
    timerfd_settime(tfd, 0, &ts, NULL);
    struct epoll_event ev = { .events = EPOLLIN, .data.ptr = current };
    epoll_ctl(epfd, EPOLL_CTL_ADD, tfd, &ev);
    coro_yield();                       /* resume when timer fires       */
    uint64_t exp;
    read(tfd, &exp, sizeof(exp));       /* drain timerfd                 */
    close(tfd);                         /* close auto-removes from epoll */
}

/* ================================================================
 * Server logic — PLAIN SEQUENTIAL CODE, just like Go!
 * ================================================================ */

static void coro_handle_client(void) {
    int fd = current->client_fd;
    set_nonblocking(fd);

    char buf[BUF_SIZE];
    ssize_t n = coro_read(fd, buf, sizeof(buf) - 1);
    if (n <= 0) { close(fd); return; }
    buf[n] = '\0';

    if (g_mode == MODE_FIB) {
        uint64_t val = 0;
        sscanf(buf, "%" SCNu64, &val);
        uint64_t result = fib(val);
        char resp[32];
        int len = snprintf(resp, sizeof(resp), "%" PRIu64 "\n", result);
        coro_write(fd, resp, len);
    } else {
        coro_sleep(SLEEP_MS);
        coro_write(fd, "42\n", 3);
    }

    close(fd);
}

/* Entry point for each coroutine — calls handle_client, then recycles */
static void coro_entry(void) {
    coro_handle_client();
    current->active = 0;
    free_stack[free_top++] = (int)(current - pool);
    coro_yield();                       /* back to scheduler, never returns */
}

/* ================================================================
 * Scheduler — the "Go runtime"
 * ================================================================ */

static void coro_spawn(int client_fd) {
    if (free_top == 0) { close(client_fd); return; }
    int idx = free_stack[--free_top];
    Coro *c = &pool[idx];
    c->active = 1;
    c->client_fd = client_fd;

    if (!c->stack_base) c->stack_base = malloc(CORO_STACK);
    /*
     * x86-64 ABI: stack must be 16-byte aligned BEFORE a `call`.
     * `call` pushes 8 bytes (return addr), so on function entry rsp % 16 == 8.
     * ctx_switch's `ret` pops 8 bytes (our fake return addr) and jumps.
     * So when coro_entry starts executing, rsp must be ≡ 8 (mod 16).
     *
     * Stack layout (grows down, top = high address):
     *   [aligned to 16]
     *   padding (8 bytes)       <- makes rsp ≡ 8 (mod 16) after ret pops entry
     *   &coro_entry             <- popped by ctx_switch's `ret`
     */
    uint64_t *sp = (uint64_t *)((uint8_t *)c->stack_base + CORO_STACK);
    sp = (uint64_t *)((uintptr_t)sp & ~15ULL);  /* align to 16 */
    *(--sp) = 0;                                 /* padding */
    *(--sp) = (uint64_t)coro_entry;              /* ctx_switch ret → here */
    c->rsp = sp;
    memset(c->regs, 0, sizeof(c->regs));

    ready_push(c);
}

/* Sentinel to distinguish listener events from coroutine events */
static Coro listener_sentinel;

static void scheduler_run(int srv) {
    set_nonblocking(srv);
    struct epoll_event ev = { .events = EPOLLIN, .data.ptr = &listener_sentinel };
    epoll_ctl(epfd, EPOLL_CTL_ADD, srv, &ev);

    struct epoll_event events[MAX_EVENTS];

    for (;;) {
        /* 1. Run all ready coroutines */
        while (rq_n > 0) {
            current = ready_pop();
            ctx_switch(&sched_ctx, current);
            current = NULL;
        }

        /* 2. Wait for I/O events */
        int nfds = epoll_wait(epfd, events, MAX_EVENTS, -1);
        if (nfds < 0) { if (errno == EINTR) continue; break; }

        for (int i = 0; i < nfds; i++) {
            Coro *c = events[i].data.ptr;

            if (c == &listener_sentinel) {
                /* Accept all pending connections, spawn coroutines */
                for (;;) {
                    struct sockaddr_in addr;
                    socklen_t len = sizeof(addr);
                    int fd = accept(srv, (struct sockaddr *)&addr, &len);
                    if (fd < 0) break;
                    coro_spawn(fd);
                }
            } else {
                /* I/O ready — resume the waiting coroutine */
                ready_push(c);
            }
        }
    }
}

/* ================================================================
 * main
 * ================================================================ */

int main(int argc, char **argv) {
    parse_mode(argc, argv);

    for (int i = 0; i < MAX_COROS; i++) {
        free_stack[i] = i;
        pool[i].stack_base = NULL;
        pool[i].active = 0;
    }
    free_top = MAX_COROS;

    epfd = epoll_create1(0);
    int srv = server_listen();

    log_msg("listening on :%d (coroutine, max %d concurrent, mode=%s)",
            PORT, MAX_COROS, argv[1]);

    scheduler_run(srv);
    return 0;
}

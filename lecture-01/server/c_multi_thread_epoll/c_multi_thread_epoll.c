#define _GNU_SOURCE

/*
 * Multi-threaded event-driven TCP server (SO_REUSEPORT + epoll).
 * Usage: ./c_multi_thread_epoll <fib|sleep>
 *
 * N threads, each with its own socket + epoll instance.
 * sleep mode uses timerfd for non-blocking sleep.
 */

#include "../c_lib/server_lib.h"
#include <fcntl.h>
#include <pthread.h>
#include <sys/epoll.h>
#include <sys/timerfd.h>

#define NUM_WORKERS 14
#define MAX_EVENTS 256

enum conn_type { CONN_LISTENER, CONN_CLIENT, CONN_TIMER };
struct conn { enum conn_type type; int fd; struct conn *peer; };

static void set_nonblocking(int fd) {
    fcntl(fd, F_SETFL, fcntl(fd, F_GETFL, 0) | O_NONBLOCK);
}

static int create_listener(void) {
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    int opt = 1;
    setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
    setsockopt(fd, SOL_SOCKET, SO_REUSEPORT, &opt, sizeof(opt));
    struct sockaddr_in addr = {
        .sin_family = AF_INET, .sin_port = htons(PORT), .sin_addr.s_addr = INADDR_ANY,
    };
    bind(fd, (struct sockaddr *)&addr, sizeof(addr));
    listen(fd, 128);
    set_nonblocking(fd);
    return fd;
}

static void *worker_thread(void *arg) {
    (void)arg;
    int srv = create_listener();
    int epfd = epoll_create1(0);

    struct conn listener = { .type = CONN_LISTENER, .fd = srv };
    struct epoll_event ev = { .events = EPOLLIN, .data.ptr = &listener };
    epoll_ctl(epfd, EPOLL_CTL_ADD, srv, &ev);

    struct epoll_event events[MAX_EVENTS];
    for (;;) {
        int nfds = epoll_wait(epfd, events, MAX_EVENTS, -1);
        if (nfds < 0) { if (errno == EINTR) continue; break; }

        for (int i = 0; i < nfds; i++) {
            struct conn *c = events[i].data.ptr;

            if (c->type == CONN_LISTENER) {
                for (;;) {
                    struct conn *cl = malloc(sizeof(*cl));
                    if (!cl) break;
                    cl->type = CONN_CLIENT; cl->peer = NULL;
                    struct sockaddr_in addr;
                    socklen_t len = sizeof(addr);
                    cl->fd = accept(srv, (struct sockaddr *)&addr, &len);
                    if (cl->fd < 0) { free(cl); break; }
                    set_nonblocking(cl->fd);
                    struct epoll_event cev = { .events = EPOLLIN, .data.ptr = cl };
                    epoll_ctl(epfd, EPOLL_CTL_ADD, cl->fd, &cev);
                }
            } else if (c->type == CONN_CLIENT) {
                char buf[BUF_SIZE];
                ssize_t n = read(c->fd, buf, sizeof(buf) - 1);
                if (n <= 0) {
                    epoll_ctl(epfd, EPOLL_CTL_DEL, c->fd, NULL);
                    close(c->fd); free(c); continue;
                }
                buf[n] = '\0';
                log_msg("req fd=%d: %s", c->fd, buf);

                if (g_mode == MODE_FIB) {
                    uint64_t val;
                    if (sscanf(buf, "%" SCNu64, &val) == 1) {
                        uint64_t result = fib(val);
                        char resp[32];
                        int len = snprintf(resp, sizeof(resp), "%" PRIu64 "\n", result);
                        log_msg("resp fd=%d: %.*s", c->fd, len - 1, resp);
                        write(c->fd, resp, len);
                    }
                    epoll_ctl(epfd, EPOLL_CTL_DEL, c->fd, NULL);
                    close(c->fd); free(c);
                } else {
                    int tfd = timerfd_create(CLOCK_MONOTONIC, TFD_NONBLOCK);
                    struct itimerspec ts = {
                        .it_value = { .tv_sec = SLEEP_MS / 1000,
                                      .tv_nsec = (SLEEP_MS % 1000) * 1000000L }
                    };
                    timerfd_settime(tfd, 0, &ts, NULL);
                    epoll_ctl(epfd, EPOLL_CTL_DEL, c->fd, NULL);

                    struct conn *tc = malloc(sizeof(*tc));
                    tc->type = CONN_TIMER; tc->fd = tfd; tc->peer = c; c->peer = tc;
                    struct epoll_event tev = { .events = EPOLLIN, .data.ptr = tc };
                    epoll_ctl(epfd, EPOLL_CTL_ADD, tfd, &tev);
                }
            } else {
                uint64_t exp;
                read(c->fd, &exp, sizeof(exp));
                struct conn *client = c->peer;
                log_msg("resp fd=%d: 42", client->fd);
                write(client->fd, "42\n", 3);
                epoll_ctl(epfd, EPOLL_CTL_DEL, c->fd, NULL);
                close(c->fd); close(client->fd);
                free(client); free(c);
            }
        }
    }
    return NULL;
}

int main(int argc, char **argv) {
    parse_mode(argc, argv);
    signal(SIGPIPE, SIG_IGN);
    log_msg("starting %d epoll workers on :%d (SO_REUSEPORT, mode=%s)", NUM_WORKERS, PORT, argv[1]);

    for (int i = 0; i < NUM_WORKERS; i++) {
        pthread_t tid;
        pthread_create(&tid, NULL, worker_thread, NULL);
        pthread_detach(tid);
    }
    for (;;) pause();
}

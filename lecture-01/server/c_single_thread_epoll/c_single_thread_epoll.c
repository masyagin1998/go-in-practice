/*
 * Single-threaded event-driven TCP server (epoll).
 * Usage: ./c_single_thread_epoll <fib|sleep>
 *
 * fib mode:   fib(N) blocks the event loop — effectively sequential.
 * sleep mode: timerfd makes the event loop truly non-blocking.
 *             One thread handles all concurrent "sleeps".
 */

#include "../c_lib/server_lib.h"
#include <fcntl.h>
#include <sys/epoll.h>
#include <sys/timerfd.h>

#define MAX_EVENTS 256

enum conn_type { CONN_LISTENER, CONN_CLIENT, CONN_TIMER };

struct conn {
    enum conn_type type;
    int fd;
    struct conn *peer;
};

static void set_nonblocking(int fd) {
    fcntl(fd, F_SETFL, fcntl(fd, F_GETFL, 0) | O_NONBLOCK);
}

int main(int argc, char **argv) {
    parse_mode(argc, argv);
    int srv = server_listen();
    set_nonblocking(srv);

    int epfd = epoll_create1(0);

    struct conn listener = { .type = CONN_LISTENER, .fd = srv };
    struct epoll_event ev = { .events = EPOLLIN, .data.ptr = &listener };
    epoll_ctl(epfd, EPOLL_CTL_ADD, srv, &ev);

    log_msg("listening on :%d (single-thread epoll, mode=%s)", PORT, argv[1]);

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
                    cl->type = CONN_CLIENT;
                    cl->peer = NULL;
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

                if (g_mode == MODE_FIB) {
                    uint64_t val;
                    if (sscanf(buf, "%" SCNu64, &val) == 1) {
                        uint64_t result = fib(val); /* blocks event loop */
                        char resp[32];
                        int len = snprintf(resp, sizeof(resp), "%" PRIu64 "\n", result);
                        write(c->fd, resp, len);
                    }
                    epoll_ctl(epfd, EPOLL_CTL_DEL, c->fd, NULL);
                    close(c->fd); free(c);
                } else {
                    /* sleep mode: create timerfd */
                    int tfd = timerfd_create(CLOCK_MONOTONIC, TFD_NONBLOCK);
                    struct itimerspec ts = {
                        .it_value = { .tv_sec = SLEEP_MS / 1000,
                                      .tv_nsec = (SLEEP_MS % 1000) * 1000000L }
                    };
                    timerfd_settime(tfd, 0, &ts, NULL);

                    epoll_ctl(epfd, EPOLL_CTL_DEL, c->fd, NULL);

                    struct conn *tc = malloc(sizeof(*tc));
                    tc->type = CONN_TIMER;
                    tc->fd = tfd;
                    tc->peer = c;
                    c->peer = tc;

                    struct epoll_event tev = { .events = EPOLLIN, .data.ptr = tc };
                    epoll_ctl(epfd, EPOLL_CTL_ADD, tfd, &tev);
                }

            } else { /* CONN_TIMER */
                uint64_t exp;
                read(c->fd, &exp, sizeof(exp));

                struct conn *client = c->peer;
                write(client->fd, "42\n", 3);

                epoll_ctl(epfd, EPOLL_CTL_DEL, c->fd, NULL);
                close(c->fd);
                close(client->fd);
                free(client); free(c);
            }
        }
    }
}

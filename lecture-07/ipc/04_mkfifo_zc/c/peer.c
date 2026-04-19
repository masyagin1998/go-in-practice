// peer.c — открывает FIFO и шлёт 5 сообщений через vmsplice(SPLICE_F_GIFT).
//
// Отличие от 02_pipes_zc: используется именованный pipe (FIFO) —
// peer и reader могут стартовать независимо, не нужен общий родитель.

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/uio.h>
#include <unistd.h>

#define FIFO_PATH  "/tmp/ipc_fifo_zc"
#define ITERATIONS 5

int main(void) {
    int fd = open(FIFO_PATH, O_WRONLY);
    if (fd < 0) { perror("open"); return 1; }

    long pagesize = sysconf(_SC_PAGESIZE);
    char *pages = aligned_alloc((size_t)pagesize, (size_t)pagesize * ITERATIONS);
    if (!pages) { perror("aligned_alloc"); return 1; }

    for (int i = 0; i < ITERATIONS; ++i) {
        char *page = pages + i * pagesize;
        int n = snprintf(page, (size_t)pagesize, "ping server %d", i);
        fprintf(stderr, "[peer %d] отправил \"%s\"\n", getpid(), page);
        page[n] = '\n';

        struct iovec iov = { .iov_base = page, .iov_len = (size_t)(n + 1) };
        if (vmsplice(fd, &iov, 1, SPLICE_F_GIFT) < 0) {
            perror("vmsplice");
            return 1;
        }
        sleep(1);
    }
    close(fd);
    return 0;
}

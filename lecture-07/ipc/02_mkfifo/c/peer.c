// peer.c — C-сервер через именованный канал (FIFO).
//
// FIFO — это файл в VFS (реально никаких блоков на диске не занимает,
// inode + буфер в ядре). Любой процесс с правом на открытие может
// присоединиться к нему без общего родителя.

#define _POSIX_C_SOURCE 200809L

#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#define FIFO_C2S   "/tmp/ipc_fifo_c2s"
#define FIFO_S2C   "/tmp/ipc_fifo_s2c"
#define BUF_SIZE   1024
#define ITERATIONS 5

int main(void) {
    // open блокируется, пока соответствующий конец не откроет Go-клиент.
    int rfd = open(FIFO_C2S, O_RDONLY);
    int wfd = open(FIFO_S2C, O_WRONLY);
    if (rfd < 0 || wfd < 0) {
        perror("open");
        return 1;
    }

    char buf[BUF_SIZE];
    for (int i = 0; i < ITERATIONS; ++i) {
        ssize_t n = read(rfd, buf, BUF_SIZE - 1);
        if (n <= 0) break;
        buf[n] = '\0';
        if (buf[n - 1] == '\n') buf[n - 1] = '\0';

        fprintf(stderr, "[peer %d] got \"%s\"\n", getpid(), buf);

        char reply[BUF_SIZE];
        int m = snprintf(reply, sizeof(reply), "echo: %s\n", buf);
        if (write(wfd, reply, (size_t)m) < 0) {
            perror("write");
            break;
        }
    }

    close(rfd);
    close(wfd);
    return 0;
}

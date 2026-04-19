// fifo peer: ping-pong через два named FIFO. argv: iters msg_size.

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define FIFO_C2S "/tmp/ipc_bench_fifo_c2s"
#define FIFO_S2C "/tmp/ipc_bench_fifo_s2c"

int main(int argc, char **argv) {
    if (argc < 3) return 1;
    int iters = atoi(argv[1]);
    size_t msg_size = (size_t)atoi(argv[2]);

    int r = open(FIFO_C2S, O_RDONLY);
    int w = open(FIFO_S2C, O_WRONLY);
    if (r < 0 || w < 0) { perror("open"); return 1; }

    char *buf = malloc(msg_size);
    if (!buf) return 1;

    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < msg_size) {
            ssize_t n = read(r, buf + got, msg_size - got);
            if (n <= 0) { free(buf); return 0; }
            got += (size_t)n;
        }
        size_t sent = 0;
        while (sent < msg_size) {
            ssize_t n = write(w, buf + sent, msg_size - sent);
            if (n <= 0) { free(buf); return 1; }
            sent += (size_t)n;
        }
    }
    free(buf);
    close(r);
    close(w);
    return 0;
}

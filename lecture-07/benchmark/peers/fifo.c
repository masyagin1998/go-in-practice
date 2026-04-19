// fifo peer: ping-pong через два named FIFO.

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define MSG_SIZE 256
#define FIFO_C2S "/tmp/ipc_bench_fifo_c2s"
#define FIFO_S2C "/tmp/ipc_bench_fifo_s2c"

int main(int argc, char **argv) {
    (void)argc;
    int iters = atoi(argv[1]);
    int r = open(FIFO_C2S, O_RDONLY);
    int w = open(FIFO_S2C, O_WRONLY);
    if (r < 0 || w < 0) { perror("open"); return 1; }

    char buf[MSG_SIZE];
    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < MSG_SIZE) {
            ssize_t n = read(r, buf + got, MSG_SIZE - got);
            if (n <= 0) return 0;
            got += (size_t)n;
        }
        size_t sent = 0;
        while (sent < MSG_SIZE) {
            ssize_t n = write(w, buf + sent, MSG_SIZE - sent);
            if (n <= 0) return 1;
            sent += (size_t)n;
        }
    }
    close(r);
    close(w);
    return 0;
}

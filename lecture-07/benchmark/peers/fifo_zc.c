// fifo_zc peer: то же, что pipe_zc, но через FIFO. argv: iters msg_size.

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/uio.h>
#include <unistd.h>

#define FIFO_C2S "/tmp/ipc_bench_fifo_zc_c2s"
#define FIFO_S2C "/tmp/ipc_bench_fifo_zc_s2c"

int main(int argc, char **argv) {
    if (argc < 3) return 1;
    int iters       = atoi(argv[1]);
    size_t msg_size = (size_t)atoi(argv[2]);

    int r = open(FIFO_C2S, O_RDONLY);
    int w = open(FIFO_S2C, O_WRONLY);
    if (r < 0 || w < 0) { perror("open"); return 1; }

    long ps = sysconf(_SC_PAGESIZE);
    size_t alloc = ((msg_size + ps - 1) / ps) * ps;
    if (alloc == 0) alloc = (size_t)ps;
    char *rx = aligned_alloc((size_t)ps, alloc);
    char *tx = aligned_alloc((size_t)ps, alloc);
    if (!rx || !tx) return 1;

    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < msg_size) {
            ssize_t n = read(r, rx + got, msg_size - got);
            if (n <= 0) goto done;
            got += (size_t)n;
        }
        size_t off = 0;
        while (off < msg_size) {
            struct iovec iov = { .iov_base = tx + off, .iov_len = msg_size - off };
            ssize_t n = vmsplice(w, &iov, 1, SPLICE_F_GIFT);
            if (n <= 0) goto done;
            off += (size_t)n;
        }
    }
done:
    (void)rx; (void)tx;
    close(r);
    close(w);
    return 0;
}

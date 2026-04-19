// pipe_zc peer: ping-pong, но на TX использует vmsplice(SPLICE_F_GIFT).
// argv: iters msg_size.
//
// Для вызова GIFT буфер должен быть page-aligned и page-sized multiple.
// Для msg_size < PAGESIZE ядро всё равно копирует, но API вызывается,
// и оверхед такого вызова тоже попадает в замер.

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/uio.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 3) return 1;
    int iters       = atoi(argv[1]);
    size_t msg_size = (size_t)atoi(argv[2]);

    long ps = sysconf(_SC_PAGESIZE);
    size_t alloc = ((msg_size + ps - 1) / ps) * ps;
    if (alloc == 0) alloc = (size_t)ps;

    char *rx  = aligned_alloc((size_t)ps, alloc);
    char *tx  = aligned_alloc((size_t)ps, alloc);
    if (!rx || !tx) return 1;

    for (int i = 0; i < iters; ++i) {
        // RX: обычный read в свой буфер (zero-copy здесь невозможен без
        // splice в другой fd).
        size_t got = 0;
        while (got < msg_size) {
            ssize_t n = read(STDIN_FILENO, rx + got, msg_size - got);
            if (n <= 0) goto done;
            got += (size_t)n;
        }
        // TX: vmsplice с SPLICE_F_GIFT. tx-страница "дарится" ядру;
        // трогать её нельзя до следующего нашего read (который вернётся
        // только после того, как reader уже её потребил).
        size_t off = 0;
        while (off < msg_size) {
            struct iovec iov = { .iov_base = tx + off, .iov_len = msg_size - off };
            ssize_t n = vmsplice(STDOUT_FILENO, &iov, 1, SPLICE_F_GIFT);
            if (n <= 0) goto done;
            off += (size_t)n;
        }
    }
done:
    // Намеренно не free() — страницы в tx могли быть "подарены" ядру,
    // и если reader их ещё не консумировал, free() = use-after-free.
    (void)rx; (void)tx;
    return 0;
}

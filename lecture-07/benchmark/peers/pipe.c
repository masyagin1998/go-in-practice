// pipe peer: ping-pong через stdin/stdout. argv: iters msg_size.

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 3) return 1;
    int iters = atoi(argv[1]);
    size_t msg_size = (size_t)atoi(argv[2]);
    char *buf = malloc(msg_size);
    if (!buf) return 1;

    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < msg_size) {
            ssize_t n = read(STDIN_FILENO, buf + got, msg_size - got);
            if (n <= 0) { free(buf); return 0; }
            got += (size_t)n;
        }
        size_t sent = 0;
        while (sent < msg_size) {
            ssize_t n = write(STDOUT_FILENO, buf + sent, msg_size - sent);
            if (n <= 0) { free(buf); return 1; }
            sent += (size_t)n;
        }
    }
    free(buf);
    return 0;
}

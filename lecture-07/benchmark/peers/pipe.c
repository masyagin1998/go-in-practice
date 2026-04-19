// pipe peer: ping-pong через stdin/stdout. argv[1] = iterations, MSG_SIZE = 256.

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define MSG_SIZE 256

int main(int argc, char **argv) {
    (void)argc;
    int iters = atoi(argv[1]);
    char buf[MSG_SIZE];

    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < MSG_SIZE) {
            ssize_t n = read(STDIN_FILENO, buf + got, MSG_SIZE - got);
            if (n <= 0) return 0;
            got += (size_t)n;
        }
        size_t sent = 0;
        while (sent < MSG_SIZE) {
            ssize_t n = write(STDOUT_FILENO, buf + sent, MSG_SIZE - sent);
            if (n <= 0) return 1;
            sent += (size_t)n;
        }
    }
    return 0;
}

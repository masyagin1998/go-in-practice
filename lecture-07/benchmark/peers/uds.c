// uds peer: подключается к AF_UNIX stream сокету. argv: iters msg_size.

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>

#define SOCK_PATH "/tmp/ipc_bench_uds"

int main(int argc, char **argv) {
    if (argc < 3) return 1;
    int iters       = atoi(argv[1]);
    size_t msg_size = (size_t)atoi(argv[2]);

    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    struct sockaddr_un addr = {0};
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, SOCK_PATH, sizeof(addr.sun_path) - 1);
    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }

    char *buf = malloc(msg_size);
    if (!buf) return 1;

    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < msg_size) {
            ssize_t n = read(fd, buf + got, msg_size - got);
            if (n <= 0) { free(buf); return 0; }
            got += (size_t)n;
        }
        size_t sent = 0;
        while (sent < msg_size) {
            ssize_t n = write(fd, buf + sent, msg_size - sent);
            if (n <= 0) { free(buf); return 1; }
            sent += (size_t)n;
        }
    }
    free(buf);
    close(fd);
    return 0;
}

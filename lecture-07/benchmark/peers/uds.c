// uds peer: подключается к AF_UNIX stream сокету.

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>

#define MSG_SIZE  256
#define SOCK_PATH "/tmp/ipc_bench_uds"

int main(int argc, char **argv) {
    (void)argc;
    int iters = atoi(argv[1]);
    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    struct sockaddr_un addr = {0};
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, SOCK_PATH, sizeof(addr.sun_path) - 1);
    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }

    char buf[MSG_SIZE];
    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < MSG_SIZE) {
            ssize_t n = read(fd, buf + got, MSG_SIZE - got);
            if (n <= 0) return 0;
            got += (size_t)n;
        }
        size_t sent = 0;
        while (sent < MSG_SIZE) {
            ssize_t n = write(fd, buf + sent, MSG_SIZE - sent);
            if (n <= 0) return 1;
            sent += (size_t)n;
        }
    }
    close(fd);
    return 0;
}

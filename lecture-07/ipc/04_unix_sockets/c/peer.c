// peer.c — C-клиент Unix domain socket (stream).

#include <stdio.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>

#define SOCK_PATH  "/tmp/ipc_unix_socket"
#define BUF_SIZE   1024
#define ITERATIONS 5

int main(void) {
    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (fd < 0) { perror("socket"); return 1; }

    struct sockaddr_un addr = {0};
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, SOCK_PATH, sizeof(addr.sun_path) - 1);

    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect");
        return 1;
    }

    char buf[BUF_SIZE];
    for (int i = 0; i < ITERATIONS; ++i) {
        int m = snprintf(buf, sizeof(buf), "ping server %d\n", i);
        if (write(fd, buf, (size_t)m) < 0) { perror("write"); break; }

        ssize_t n = read(fd, buf, BUF_SIZE - 1);
        if (n <= 0) break;
        buf[n] = '\0';
        if (buf[n - 1] == '\n') buf[n - 1] = '\0';
        fprintf(stderr, "[peer %d] got \"%s\"\n", getpid(), buf);
    }

    close(fd);
    return 0;
}

// peer.c — C-клиент TCP: подключается к Go-серверу на 127.0.0.1.

#include <arpa/inet.h>
#include <stdio.h>
#include <string.h>
#include <sys/socket.h>
#include <unistd.h>

#define PORT       8891
#define BUF_SIZE   1024
#define ITERATIONS 5

int main(void) {
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) { perror("socket"); return 1; }

    struct sockaddr_in addr = {0};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(PORT);
    inet_pton(AF_INET, "127.0.0.1", &addr.sin_addr);

    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect");
        return 1;
    }

    char buf[BUF_SIZE];
    for (int i = 0; i < ITERATIONS; ++i) {
        char msg[64];
        int mlen = snprintf(msg, sizeof(msg), "ping server %d", i);
        fprintf(stderr, "[peer %d] отправил \"%s\"\n", getpid(), msg);
        msg[mlen] = '\n';
        if (write(fd, msg, (size_t)(mlen + 1)) < 0) { perror("write"); break; }

        ssize_t n = read(fd, buf, BUF_SIZE - 1);
        if (n <= 0) break;
        buf[n] = '\0';
        if (buf[n - 1] == '\n') buf[n - 1] = '\0';
        fprintf(stderr, "[peer %d] получил \"%s\"\n", getpid(), buf);
    }

    close(fd);
    return 0;
}

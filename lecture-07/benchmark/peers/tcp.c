// tcp peer: подключается к 127.0.0.1:PORT, ping-pong. argv: iters port msg_size.

#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 4) { fprintf(stderr, "usage: %s <iters> <port> <msg_size>\n", argv[0]); return 1; }
    int iters       = atoi(argv[1]);
    int port        = atoi(argv[2]);
    size_t msg_size = (size_t)atoi(argv[3]);

    int fd = socket(AF_INET, SOCK_STREAM, 0);
    struct sockaddr_in addr = {0};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    inet_pton(AF_INET, "127.0.0.1", &addr.sin_addr);
    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }
    int one = 1;
    setsockopt(fd, IPPROTO_TCP, 1 /* TCP_NODELAY */, &one, sizeof(one));

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

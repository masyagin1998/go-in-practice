// tcp peer: подключается к 127.0.0.1:PORT, ping-pong.

#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <unistd.h>

#define MSG_SIZE 256

int main(int argc, char **argv) {
    if (argc < 3) { fprintf(stderr, "usage: %s <iters> <port>\n", argv[0]); return 1; }
    int iters = atoi(argv[1]);
    int port  = atoi(argv[2]);
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    struct sockaddr_in addr = {0};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    inet_pton(AF_INET, "127.0.0.1", &addr.sin_addr);
    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }
    // отключаем Nagle для честного low-latency ping-pong
    int one = 1;
    setsockopt(fd, IPPROTO_TCP, 1 /* TCP_NODELAY */, &one, sizeof(one));

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

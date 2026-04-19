// tcp_zc peer: TCP + MSG_ZEROCOPY. argv: iters port msg_size.
//
// Каждый send(MSG_ZEROCOPY) добавляет запись в error-queue; если её не
// дренировать, очередь переполняется и send начинает возвращать ENOBUFS.
// Дренируем раз в N итераций одним recvmsg (ядро отдаёт все pending
// уведомления одним пакетом).

#include <arpa/inet.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <unistd.h>

#ifndef MSG_ZEROCOPY
#define MSG_ZEROCOPY 0x4000000
#endif
#ifndef SO_ZEROCOPY
#define SO_ZEROCOPY 60
#endif

#define DRAIN_EVERY 64

static void drain(int fd) {
    char ctl[1024];
    struct msghdr m;
    for (;;) {
        memset(&m, 0, sizeof(m));
        m.msg_control = ctl;
        m.msg_controllen = sizeof(ctl);
        if (recvmsg(fd, &m, MSG_ERRQUEUE | MSG_DONTWAIT) < 0) return;
    }
}

int main(int argc, char **argv) {
    if (argc < 4) { fprintf(stderr, "usage: %s <iters> <port> <msg_size>\n", argv[0]); return 1; }
    int iters       = atoi(argv[1]);
    int port        = atoi(argv[2]);
    size_t msg_size = (size_t)atoi(argv[3]);

    int fd = socket(AF_INET, SOCK_STREAM, 0);
    int one = 1;
    setsockopt(fd, IPPROTO_TCP, 1 /* TCP_NODELAY */, &one, sizeof(one));
    if (setsockopt(fd, SOL_SOCKET, SO_ZEROCOPY, &one, sizeof(one)) < 0) {
        perror("SO_ZEROCOPY"); return 1;
    }

    struct sockaddr_in addr = {0};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    inet_pton(AF_INET, "127.0.0.1", &addr.sin_addr);
    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }

    char *rx = malloc(msg_size);
    char *tx = malloc(msg_size);
    if (!rx || !tx) return 1;

    for (int i = 0; i < iters; ++i) {
        size_t got = 0;
        while (got < msg_size) {
            ssize_t n = read(fd, rx + got, msg_size - got);
            if (n <= 0) goto done;
            got += (size_t)n;
        }
        size_t off = 0;
        while (off < msg_size) {
            ssize_t n = send(fd, tx + off, msg_size - off, MSG_ZEROCOPY);
            if (n <= 0) {
                if (errno == ENOBUFS) { drain(fd); continue; }
                goto done;
            }
            off += (size_t)n;
        }
        if ((i & (DRAIN_EVERY - 1)) == 0) drain(fd);
    }
done:
    free(rx); free(tx);
    close(fd);
    return 0;
}

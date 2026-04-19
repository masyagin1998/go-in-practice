// peer.c — TCP-клиент, шлёт через send(..., MSG_ZEROCOPY).
//
// Этапы:
// 1. setsockopt(SO_ZEROCOPY) включает режим для сокета.
// 2. send(..., MSG_ZEROCOPY) — ядро не копирует буфер, а пинит
//    его страницы и отдаёт TCP-стеку как скаттер.
// 3. Когда данные реально уехали и буфер можно переиспользовать,
//    ядро кладёт уведомление в error-queue (MSG_ERRQUEUE).
//    Без дренирования error-queue она переполнится и send начнёт
//    падать с ENOBUFS.

#include <arpa/inet.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <unistd.h>

#define PORT       8892
#define BUF_SIZE   65536
#define ITERATIONS 5

#ifndef MSG_ZEROCOPY
#define MSG_ZEROCOPY 0x4000000
#endif

#ifndef SO_ZEROCOPY
#define SO_ZEROCOPY 60
#endif

// Сливаем error-queue — уведомления о том, что отправленные через
// MSG_ZEROCOPY буферы уже уехали и их можно переиспользовать.
static void drain_errqueue(int fd) {
    char ctl[256];
    struct msghdr msg;
    for (;;) {
        memset(&msg, 0, sizeof(msg));
        msg.msg_control = ctl;
        msg.msg_controllen = sizeof(ctl);
        if (recvmsg(fd, &msg, MSG_ERRQUEUE | MSG_DONTWAIT) < 0) {
            if (errno == EAGAIN || errno == EWOULDBLOCK) return;
            perror("recvmsg ERRQUEUE");
            return;
        }
    }
}

int main(void) {
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) { perror("socket"); return 1; }

    int on = 1;
    if (setsockopt(fd, SOL_SOCKET, SO_ZEROCOPY, &on, sizeof(on)) < 0) {
        perror("setsockopt(SO_ZEROCOPY)");
        return 1;
    }

    struct sockaddr_in addr = {0};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(PORT);
    inet_pton(AF_INET, "127.0.0.1", &addr.sin_addr);
    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }

    char *buf = calloc(1, BUF_SIZE);
    if (!buf) { perror("calloc"); return 1; }

    for (int i = 0; i < ITERATIONS; ++i) {
        int n = snprintf(buf, BUF_SIZE, "ping server %d", i);
        fprintf(stderr, "[peer %d] отправил \"%s\"\n", getpid(), buf);
        buf[n] = '\n';

        if (send(fd, buf, (size_t)(n + 1), MSG_ZEROCOPY) < 0) {
            perror("send"); break;
        }

        char reply[BUF_SIZE];
        ssize_t r = read(fd, reply, BUF_SIZE - 1);
        if (r <= 0) break;
        reply[r] = '\0';
        if (reply[r - 1] == '\n') reply[r - 1] = '\0';
        fprintf(stderr, "[peer %d] получил \"%s\"\n", getpid(), reply);

        // После того как reply пришёл, TX точно завершился → дренируем.
        drain_errqueue(fd);
    }

    free(buf);
    close(fd);
    return 0;
}

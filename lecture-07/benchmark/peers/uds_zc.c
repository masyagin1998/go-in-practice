// uds_zc peer: SCM_RIGHTS + memfd + mmap. argv: iters msg_size.
//
// Протокол:
// 1. peer создаёт memfd размером 2*msg_size, mmap'ит, отсылает fd Go.
// 2. Layout: [0..msg_size] — go→peer, [msg_size..2*msg_size] — peer→go.
// 3. Синхронизация — 1-байтовые сообщения на том же Unix-сокете.
//    На каждой итерации: Go пишет в go→peer, шлёт байт; peer читает
//    байт, копирует echo в peer→go, шлёт байт; Go читает байт.
//
// Данные никогда не ходят через сокет — только 1-байтовые сигналы.

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/socket.h>
#include <sys/syscall.h>
#include <sys/un.h>
#include <unistd.h>

#define SOCK_PATH "/tmp/ipc_bench_uds_zc"

static int send_fd(int sock, int fd) {
    char dummy = 0;
    struct iovec iov = { .iov_base = &dummy, .iov_len = 1 };
    char ctl[CMSG_SPACE(sizeof(int))] = {0};
    struct msghdr msg = {0};
    msg.msg_iov = &iov;
    msg.msg_iovlen = 1;
    msg.msg_control = ctl;
    msg.msg_controllen = sizeof(ctl);
    struct cmsghdr *c = CMSG_FIRSTHDR(&msg);
    c->cmsg_level = SOL_SOCKET;
    c->cmsg_type  = SCM_RIGHTS;
    c->cmsg_len   = CMSG_LEN(sizeof(int));
    *((int *)CMSG_DATA(c)) = fd;
    return (int)sendmsg(sock, &msg, 0);
}

int main(int argc, char **argv) {
    if (argc < 3) return 1;
    int iters       = atoi(argv[1]);
    size_t msg_size = (size_t)atoi(argv[2]);
    size_t total    = msg_size * 2;

    int sock = socket(AF_UNIX, SOCK_STREAM, 0);
    struct sockaddr_un addr = {0};
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, SOCK_PATH, sizeof(addr.sun_path) - 1);
    if (connect(sock, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }

    int mfd = memfd_create("bench", 0);
    if (mfd < 0 || ftruncate(mfd, (off_t)total) < 0) { perror("memfd"); return 1; }

    char *shared = mmap(NULL, total, PROT_READ | PROT_WRITE, MAP_SHARED, mfd, 0);
    if (shared == MAP_FAILED) { perror("mmap"); return 1; }

    if (send_fd(sock, mfd) < 0) { perror("send_fd"); return 1; }
    close(mfd); // у Go уже есть копия

    char *in  = shared;
    char *out = shared + msg_size;
    char sig;

    for (int i = 0; i < iters; ++i) {
        if (read(sock, &sig, 1) <= 0) break;
        memcpy(out, in, msg_size); // echo
        if (write(sock, &sig, 1) <= 0) break;
    }

    munmap(shared, total);
    close(sock);
    return 0;
}

// peer.c — создаёт memfd, пишет туда строку, передаёт fd через SCM_RIGHTS.
//
// memfd_create(2) возвращает fd, указывающий на анонимный файл в tmpfs
// (физически — страницы в page cache). Через AF_UNIX + SCM_RIGHTS fd
// можно отдать другому процессу; ядро дублирует запись в его fd-таблицу.

#include <stdio.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/socket.h>
#include <sys/syscall.h>
#include <sys/un.h>
#include <unistd.h>

#define SOCK_PATH  "/tmp/ipc_uds_zc"
#define ITERATIONS 5

static int send_fd(int sock, int fd_to_send) {
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
    *((int *)CMSG_DATA(c)) = fd_to_send;

    return (int)sendmsg(sock, &msg, 0);
}

int main(void) {
    int sock = socket(AF_UNIX, SOCK_STREAM, 0);
    if (sock < 0) { perror("socket"); return 1; }

    struct sockaddr_un addr = {0};
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, SOCK_PATH, sizeof(addr.sun_path) - 1);
    if (connect(sock, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("connect"); return 1;
    }

    for (int i = 0; i < ITERATIONS; ++i) {
        int mfd = memfd_create("ping", 0);
        if (mfd < 0) { perror("memfd_create"); return 1; }

        char msg[64];
        int n = snprintf(msg, sizeof(msg), "ping server %d", i);
        if (write(mfd, msg, (size_t)n) != n) { perror("write"); return 1; }
        // Обрезаем до длины данных, чтобы получатель mmap'ил ровно их.
        if (ftruncate(mfd, (off_t)n) < 0) { perror("ftruncate"); return 1; }

        if (send_fd(sock, mfd) < 0) { perror("sendmsg"); return 1; }
        fprintf(stderr, "[peer %d] отправил \"%s\"\n", getpid(), msg);
        close(mfd); // у Go уже есть своя копия fd
        usleep(100000);
    }

    close(sock);
    return 0;
}

// peer.c — шлёт N сигналов SIGUSR1 Go-процессу.
// Использование: ./peer <pid> <count>

#define _XOPEN_SOURCE 500

#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 3) { fprintf(stderr, "usage: %s <pid> <count>\n", argv[0]); return 1; }
    pid_t pid = (pid_t)atoi(argv[1]);
    int n = atoi(argv[2]);

    for (int i = 0; i < n; ++i) {
        if (kill(pid, SIGUSR1) < 0) { perror("kill"); return 1; }
        usleep(5000); // 5ms между сигналами
    }
    fprintf(stderr, "[peer] отправлено %d SIGUSR1\n", n);
    return 0;
}

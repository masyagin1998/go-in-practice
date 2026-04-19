// peer.c — шлёт Go-родителю пять sigqueue'ов с int-payload'ом.
//
// SIGUSR1/SIGUSR2 чередуются; значение 100+i. Go читает их через
// signalfd и показывает полученный signo + ssi_int + ssi_pid.

#define _GNU_SOURCE

#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 2) { fprintf(stderr, "usage: %s <parent_pid>\n", argv[0]); return 1; }
    pid_t parent = (pid_t)atoi(argv[1]);

    for (int i = 0; i < 5; ++i) {
        union sigval v;
        v.sival_int = 100 + i;
        int signo = (i % 2 == 0) ? SIGUSR1 : SIGUSR2;
        if (sigqueue(parent, signo, v) < 0) {
            perror("sigqueue");
            return 1;
        }
        fprintf(stderr, "[peer %d] sigqueue(%d, SIGUSR%d, val=%d)\n",
                getpid(), parent, (i % 2 == 0) ? 1 : 2, v.sival_int);
        usleep(150000);
    }
    return 0;
}

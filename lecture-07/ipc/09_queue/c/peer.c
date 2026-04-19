// peer.c — C-сервер через POSIX message queue.
//
// POSIX mq_* отличается от pipe/socket тем, что сохраняет границы
// сообщений и поддерживает приоритеты (0..sysconf(_SC_MQ_PRIO_MAX)).
// Очередь живёт в /dev/mqueue.

#include "shared.h"

#include <fcntl.h>
#include <mqueue.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

int main(void) {
    mqd_t c2s = mq_open(MQ_C2S, O_RDONLY);
    mqd_t s2c = mq_open(MQ_S2C, O_WRONLY);
    if (c2s == (mqd_t)-1 || s2c == (mqd_t)-1) { perror("mq_open"); return 1; }

    char buf[MQ_MSGSIZE];
    for (int i = 0; i < ITERATIONS; ++i) {
        unsigned prio = 0;
        ssize_t n = mq_receive(c2s, buf, MQ_MSGSIZE, &prio);
        if (n <= 0) break;
        buf[n] = '\0';
        fprintf(stderr, "[peer %d] получил \"%s\" (prio=%u)\n", getpid(), buf, prio);

        char reply[MQ_MSGSIZE];
        int m = snprintf(reply, sizeof(reply), "echo: %.900s", buf);
        fprintf(stderr, "[peer %d] отправил \"%s\" (prio=%u)\n", getpid(), reply, prio);
        mq_send(s2c, reply, (size_t)m, prio);
    }

    mq_close(c2s);
    mq_close(s2c);
    return 0;
}

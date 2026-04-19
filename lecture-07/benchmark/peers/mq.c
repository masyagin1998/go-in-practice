// mq peer: ping-pong через POSIX message queue.

#include <fcntl.h>
#include <mqueue.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MSG_SIZE 256
#define MQ_C2S "/ipc_bench_mq_c2s"
#define MQ_S2C "/ipc_bench_mq_s2c"

int main(int argc, char **argv) {
    (void)argc;
    int iters = atoi(argv[1]);
    mqd_t c2s = mq_open(MQ_C2S, O_RDONLY);
    mqd_t s2c = mq_open(MQ_S2C, O_WRONLY);
    if (c2s == (mqd_t)-1 || s2c == (mqd_t)-1) { perror("mq_open"); return 1; }

    char buf[MSG_SIZE];
    for (int i = 0; i < iters; ++i) {
        ssize_t n = mq_receive(c2s, buf, MSG_SIZE, NULL);
        if (n <= 0) return 1;
        if (mq_send(s2c, buf, MSG_SIZE, 0) < 0) return 1;
    }
    mq_close(c2s);
    mq_close(s2c);
    return 0;
}

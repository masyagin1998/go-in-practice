// peer.c — пишет 5 сообщений в stdout. shell pipe доставит их в stdin Go.
//
// Запуск: ./peer | go run .

#include <stdio.h>
#include <unistd.h>

#define ITERATIONS 5

int main(void) {
    for (int i = 0; i < ITERATIONS; ++i) {
        char msg[64];
        snprintf(msg, sizeof(msg), "ping server %d", i);
        fprintf(stderr, "[peer %d] отправил \"%s\"\n", getpid(), msg);
        printf("%s\n", msg);
        fflush(stdout); // без flush Go ничего не увидит до exit'а
        sleep(1);       // имитация "долгой работы" на стороне C
    }
    return 0;
}

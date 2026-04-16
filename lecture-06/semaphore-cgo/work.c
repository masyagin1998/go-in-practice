// work.c — busy-wait на заданное количество миллисекунд.
#include "work.h"
#include <time.h>

void cpu_work(int ms) {
    struct timespec start, now;
    clock_gettime(CLOCK_MONOTONIC, &start);

    long target_ns = (long)ms * 1000000L;
    for (;;) {
        clock_gettime(CLOCK_MONOTONIC, &now);
        long elapsed = (now.tv_sec - start.tv_sec) * 1000000000L
                     + (now.tv_nsec - start.tv_nsec);
        if (elapsed >= target_ns) {
            break;
        }
    }
}

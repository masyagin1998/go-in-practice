/*
 * Stress test: пытаемся запустить 1 000 000 OS-тредов, каждый спит 10 секунд.
 * Демонстрирует реальную стоимость pthread и лимиты ядра на количество тредов.
 *
 * Перед запуском может потребоваться:
 *   sudo sysctl -w vm.max_map_count=4200000
 *   sudo sysctl -w kernel.threads-max=2100000
 *   sudo sysctl -w kernel.pid_max=2100000
 *
 * Следить: watch -n1 'ps -o pid,vsz,rss,nlwp -p $(pgrep stress_test_c)'
 *
 * gcc -O2 -Wall -pthread -o stress_test_c stress_test.c
 */

#define _GNU_SOURCE
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#define NUM_THREADS  1000000
#define SLEEP_SEC    10
static void *thread_func(void *arg) {
    (void)arg;
    sleep(SLEEP_SEC);
    return NULL;
}

static double now_sec(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts.tv_sec + ts.tv_nsec / 1e9;
}

static void print_proc_status(void) {
    FILE *f = fopen("/proc/self/status", "r");
    if (!f) return;
    char line[256];
    while (fgets(line, sizeof(line), f)) {
        if (strncmp(line, "Threads", 7) == 0) {
            printf("%s", line);
        } else if (strncmp(line, "VmSize", 6) == 0 ||
                   strncmp(line, "VmRSS",  5) == 0) {
            char name[32];
            long kb;
            if (sscanf(line, "%31[^:]: %ld", name, &kb) == 2)
                printf("%s:\t%ld MB\n", name, kb / 1024);
        }
    }
    fclose(f);
}

int main(void) {
    printf("Запускаем %d тредов (sleep %d сек, стек по умолчанию ~8 MB)...\n",
           NUM_THREADS, SLEEP_SEC);

    pthread_t *tids = malloc(sizeof(pthread_t) * NUM_THREADS);
    if (!tids) { perror("malloc"); return 1; }

    double t0 = now_sec();
    int created = 0;

    for (int i = 0; i < NUM_THREADS; i++) {
        int rc = pthread_create(&tids[i], NULL, thread_func, NULL);
        if (rc != 0) {
            fprintf(stderr, "pthread_create failed at thread %d: %s\n", i, strerror(rc));
            break;
        }
        created++;
        if (created % 50000 == 0) {
            printf("  создано %d тредов (%.1f сек)...\n", created, now_sec() - t0);
        }
    }

    double launched = now_sec() - t0;
    printf("Создано %d тредов за %.2f сек\n", created, launched);

    print_proc_status();

    printf("Ждём завершения...\n");
    for (int i = 0; i < created; i++) {
        pthread_join(tids[i], NULL);
    }

    printf("Готово за %.2f сек\n", now_sec() - t0);

    free(tids);
    return 0;
}

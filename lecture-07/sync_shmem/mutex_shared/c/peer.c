// peer.c — C-воркер. Получает id через argv, открывает общий shmem,
// в цикле берёт mutex, инкрементит counter, печатает сообщение и отпускает.
//
// Интерес: pthread_mutex с атрибутом PTHREAD_PROCESS_SHARED корректно
// работает между **разными** процессами, если лежит в shared-mapping.

#define _XOPEN_SOURCE 500

#include "shared.h"

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 2) { fprintf(stderr, "usage: %s <id>\n", argv[0]); return 1; }
    int id = atoi(argv[1]);

    int fd = shm_open(SHM_NAME, O_RDWR, 0);
    if (fd < 0) { perror("shm_open"); return 1; }
    shared_data_t *d = mmap(NULL, sizeof(*d), PROT_READ | PROT_WRITE,
                            MAP_SHARED, fd, 0);
    close(fd);
    if (d == MAP_FAILED) { perror("mmap"); return 1; }

    for (int i = 0; i < ITERATIONS; ++i) {
        pthread_mutex_lock(&d->mutex);
        d->counter++;
        fprintf(stderr, "worker %d pid=%d counter=%d\n", id, getpid(), d->counter);
        usleep(1000);
        pthread_mutex_unlock(&d->mutex);
    }

    munmap(d, sizeof(*d));
    return 0;
}

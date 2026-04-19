// peer.c — C-сервер через shmem + POSIX-семафоры.
//
// Вместо busy-wait из shmem/ ждём на sem_wait(sem_client). После ответа
// делаем sem_post(sem_server) — Go-клиент просыпается. Стоимость ожидания:
// ~0 CPU (ждём в ядре).

#include "shared.h"

#include <fcntl.h>
#include <semaphore.h>
#include <stdio.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

int main(void) {
    int fd = shm_open(SHM_NAME, O_RDWR, 0);
    if (fd < 0) { perror("shm_open"); return 1; }
    shared_mem_t *shm = mmap(NULL, sizeof(*shm), PROT_READ | PROT_WRITE,
                             MAP_SHARED, fd, 0);
    close(fd);
    if (shm == MAP_FAILED) { perror("mmap"); return 1; }

    sem_t *sc = sem_open(SEM_CLIENT, 0);
    sem_t *ss = sem_open(SEM_SERVER, 0);
    if (sc == SEM_FAILED || ss == SEM_FAILED) { perror("sem_open"); return 1; }

    for (int i = 0; i < ITERATIONS; ++i) {
        sem_wait(sc);

        char temp[SHM_BUF];
        strncpy(temp, shm->buffer, SHM_BUF - 1);
        temp[SHM_BUF - 1] = '\0';
        fprintf(stderr, "[peer %d] got \"%s\"\n", getpid(), temp);

        snprintf(shm->buffer, SHM_BUF, "echo: %.900s", temp);
        sem_post(ss);
    }

    sem_close(sc);
    sem_close(ss);
    munmap(shm, sizeof(*shm));
    return 0;
}

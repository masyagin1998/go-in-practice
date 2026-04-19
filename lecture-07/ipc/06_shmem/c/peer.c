// peer.c — C-сервер через named shared memory.
//
// Открываем ту же /ipc_shmem_demo, mmap'им, в цикле ждём client_ready,
// пишем echo, выставляем server_ready. Синхронизация — busy-wait через
// volatile sig_atomic_t. Это наивно и плохо для CPU; в соседнем примере
// semaphore/ заменим на POSIX-семафоры.

#define _XOPEN_SOURCE 500

#include "shared.h"

#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

int main(void) {
    int fd = shm_open(SHM_NAME, O_RDWR, 0);
    if (fd < 0) { perror("shm_open"); return 1; }

    shared_mem_t *shm = mmap(NULL, sizeof(*shm), PROT_READ | PROT_WRITE,
                             MAP_SHARED, fd, 0);
    if (shm == MAP_FAILED) { perror("mmap"); return 1; }
    close(fd);

    for (int i = 0; i < ITERATIONS; ++i) {
        while (!shm->client_ready) usleep(100);

        char temp[SHM_BUF];
        strncpy(temp, shm->buffer, SHM_BUF - 1);
        temp[SHM_BUF - 1] = '\0';
        fprintf(stderr, "[peer %d] got \"%s\"\n", getpid(), temp);

        snprintf(shm->buffer, SHM_BUF, "echo: %.900s", temp);
        shm->server_ready = 1;
        shm->client_ready = 0;
    }

    munmap(shm, sizeof(*shm));
    return 0;
}

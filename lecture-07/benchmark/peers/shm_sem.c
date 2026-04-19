// shm+sem peer: общий shmem + два именованных POSIX-семафора. argv: iters msg_size.

#include <fcntl.h>
#include <semaphore.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

#define SHM_NAME   "/ipc_bench_shm"
#define SEM_CLIENT "/ipc_bench_sem_c"
#define SEM_SERVER "/ipc_bench_sem_s"

int main(int argc, char **argv) {
    if (argc < 3) return 1;
    int iters       = atoi(argv[1]);
    size_t msg_size = (size_t)atoi(argv[2]);

    int fd = shm_open(SHM_NAME, O_RDWR, 0);
    if (fd < 0) { perror("shm_open"); return 1; }
    char *buf = mmap(NULL, msg_size, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    close(fd);
    if (buf == MAP_FAILED) return 1;

    sem_t *sc = sem_open(SEM_CLIENT, 0);
    sem_t *ss = sem_open(SEM_SERVER, 0);
    if (sc == SEM_FAILED || ss == SEM_FAILED) return 1;

    // Честное сравнение с pipe/fifo/tcp/uds: в тех транспортах на каждой
    // итерации копируется msg_size байт в обе стороны. Здесь эмулируем
    // "echo": забираем сообщение во временный массив и возвращаем его
    // обратно. Без этого бенчмарк измерял бы только sem_wait/post,
    // игнорируя стоимость копирования через общую память.
    char *local = malloc(msg_size);
    if (!local) return 1;
    for (int i = 0; i < iters; ++i) {
        sem_wait(sc);
        memcpy(local, buf, msg_size);
        memcpy(buf, local, msg_size);
        sem_post(ss);
    }
    free(local);
    sem_close(sc);
    sem_close(ss);
    munmap(buf, msg_size);
    return 0;
}

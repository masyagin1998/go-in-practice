// shm+sem peer: общий shmem + два именованных POSIX-семафора.

#include <fcntl.h>
#include <semaphore.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

#define MSG_SIZE 256
#define SHM_NAME   "/ipc_bench_shm"
#define SEM_CLIENT "/ipc_bench_sem_c"
#define SEM_SERVER "/ipc_bench_sem_s"

int main(int argc, char **argv) {
    (void)argc;
    int iters = atoi(argv[1]);

    int fd = shm_open(SHM_NAME, O_RDWR, 0);
    if (fd < 0) { perror("shm_open"); return 1; }
    char *buf = mmap(NULL, MSG_SIZE, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    close(fd);
    if (buf == MAP_FAILED) return 1;

    sem_t *sc = sem_open(SEM_CLIENT, 0);
    sem_t *ss = sem_open(SEM_SERVER, 0);
    if (sc == SEM_FAILED || ss == SEM_FAILED) return 1;

    // Честное сравнение с pipe/fifo/tcp/uds: в тех транспортах на каждой
    // итерации копируется MSG_SIZE байт в обе стороны. Здесь эмулируем
    // "echo": забираем сообщение во временный массив и возвращаем его
    // обратно. Без этого бенчмарк измерял бы только sem_wait/post,
    // игнорируя стоимость копирования через общую память.
    char local[MSG_SIZE];
    for (int i = 0; i < iters; ++i) {
        sem_wait(sc);
        memcpy(local, buf, MSG_SIZE);
        memcpy(buf, local, MSG_SIZE);
        sem_post(ss);
    }
    sem_close(sc);
    sem_close(ss);
    munmap(buf, MSG_SIZE);
    return 0;
}

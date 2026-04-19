// peer.c — C-дочка. Открывает общий robust-mutex, захватывает его,
// пишет значение и выходит БЕЗ разблокировки. Эмулируем падение
// воркера с удержанием критической секции.

#include "shared.h"

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <unistd.h>

int main(void) {
    int fd = shm_open(SHM_NAME, O_RDWR, 0);
    if (fd < 0) { perror("shm_open"); return 1; }

    shared_t *s = mmap(NULL, sizeof(*s), PROT_READ | PROT_WRITE,
                       MAP_SHARED, fd, 0);
    close(fd);
    if (s == MAP_FAILED) { perror("mmap"); return 1; }

    if (pthread_mutex_lock(&s->mutex) != 0) {
        fprintf(stderr, "[child] lock failed\n");
        return 1;
    }
    s->value = 42;
    fprintf(stderr, "[child pid=%d] locked, value=42 — падаем без unlock\n",
            getpid());

    // _exit — без atexit/flush, чтобы не дёргать деструкторы.
    _exit(0);
}

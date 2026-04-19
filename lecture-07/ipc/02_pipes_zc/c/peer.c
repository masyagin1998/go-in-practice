// peer.c — пишет 5 сообщений в pipe через vmsplice(SPLICE_F_GIFT).
//
// Обычный write(2) копирует байты user→kernel в pipe-буфер. vmsplice
// с SPLICE_F_GIFT передаёт ядру ссылки на user-страницы: после вызова
// пользователь обещает не трогать память, ядро может её забрать.
//
// Ограничения:
// - работает только если write-end — pipe/FIFO (shell pipe сюда подходит);
// - реально zero-copy срабатывает только для page-aligned, page-sized
//   кусков. Для мелких сообщений ядро всё равно копирует.

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/uio.h>
#include <unistd.h>

#define ITERATIONS 5

int main(void) {
    long pagesize = sysconf(_SC_PAGESIZE);
    // Выделяем по странице на сообщение — чтобы каждый вызов vmsplice
    // имел шанс реально подарить страницу ядру (SPLICE_F_GIFT любит
    // page-aligned). Если меньше — ядро копирует.
    char *pages = aligned_alloc((size_t)pagesize, (size_t)pagesize * ITERATIONS);
    if (!pages) { perror("aligned_alloc"); return 1; }

    for (int i = 0; i < ITERATIONS; ++i) {
        char *page = pages + i * pagesize;
        int n = snprintf(page, (size_t)pagesize, "ping server %d", i);
        fprintf(stderr, "[peer %d] отправил \"%s\"\n", getpid(), page);
        page[n] = '\n';

        struct iovec iov = { .iov_base = page, .iov_len = (size_t)(n + 1) };
        if (vmsplice(STDOUT_FILENO, &iov, 1, SPLICE_F_GIFT) < 0) {
            perror("vmsplice");
            return 1;
        }
        sleep(1);
    }
    // Страницу не освобождаем: мы "подарили" её ядру, трогать нельзя
    // до того, как Go её прочитает. Процесс всё равно завершается.
    return 0;
}

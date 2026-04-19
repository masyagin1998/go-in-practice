# 06_shmem — named POSIX shared memory

`shm_open(3)` → fd на область в tmpfs (`/dev/shm/`). Оба процесса
mmap'ят одну память. Самый быстрый IPC: `memcpy`, без syscall'ов
на горячем пути.

Синхронизация здесь — примитивный busy-wait через `atomic.Int32` /
`volatile sig_atomic_t`. Это 100% CPU пока ждём. Правильный способ
рядом в `../07_semaphore/`.

## Запуск

```bash
make peer
# терминал 1: go run .
# терминал 2: ./peer
```

или `make run`.

## Реальные кейсы

- **High-throughput data path**: LMAX Disruptor, Aeron, DPDK — кадры
  в shmem, обход kernel.
- **ML inference**: Triton/TFServing пишут тензоры в shmem, клиенты
  читают без копирования.
- **Chrome**: renderer ↔ GPU ↔ browser — видео и текстуры через shmem.
- **Postgres `shared_buffers`** — общий кеш страниц между backend'ами.

## Про анонимный shmem (`MAP_ANONYMOUS | MAP_SHARED`)

В C-курсе был `shmem_anon.c` — shmem до `fork()`, наследуется через VMA.
Между неродственными процессами не работает (VMA не переживает exec).
Альтернатива — `memfd_create` + `SCM_RIGHTS` по UDS (отдельная тема).

## Подводные камни

- **Memory ordering**: `volatile sig_atomic_t` НЕ даёт acquire/release
  на ARM/Power. В проде — `stdatomic.h` или semaphore/mutex.
- **Cleanup**: `shm_unlink` убирает имя, но память жива пока процессы
  не сделают `munmap`.
- **Размер**: tmpfs ограничен `/proc/sys/kernel/shmmax` и свободной RAM.

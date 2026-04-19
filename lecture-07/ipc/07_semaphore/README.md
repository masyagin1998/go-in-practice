# 07_semaphore — shmem + POSIX named semaphore

В `06_shmem` мы ждали флаги busy-wait'ом — CPU горел. Здесь два
named-семафора: `sem_wait` блокирует поток в ядре (futex), пока партнёр
не сделает `sem_post`. Zero CPU в ожидании.

Семафор виден как `/dev/shm/sem.<name>`.

## Запуск

```bash
make peer
# терминал 1: go run .
# терминал 2: ./peer
```

или `make run`.

## Реальные кейсы

- **Rate limiting** между процессами: N GPU-слотов — `sem_trywait` до запуска,
  `sem_post` после; `EAGAIN` → воркер ждёт.
- **Producer/consumer** над кольцевым буфером в shmem: два семафора —
  "свободные" и "занятые слоты".
- **Rendezvous между неродственными процессами** — когда `pthread_mutex`
  недоступен (разные runtime'ы).

## Альтернативы

- **`eventfd` + epoll** — Linux, удобнее интегрируется с event-loop.
- **UDS datagram** — для редких событий проще и гибче.
- **Futex напрямую** — быстрее, но low-level.

## Подводные камни

- **Stuck semaphore**: владелец упал между `sem_wait` и `sem_post` →
  семафор заклинил. В отличие от `pthread_mutex_robust`, POSIX sem не
  имеет recovery. Для критических секций — robust mutex
  (см. `../../sync_shmem/mutex_death/`).
- **Cleanup**: `sem_unlink` убирает имя; процесс держит открытый
  дескриптор. Забытые остаются в `/dev/shm/sem.*`.

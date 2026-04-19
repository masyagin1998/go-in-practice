# ipc/ — способы обмена между процессами

Каждый пример — два процесса (Go и C-peer), которые обмениваются
сообщениями через конкретный IPC-механизм. Пары `*_zc` — zero-copy
вариант того же транспорта.

## Карта

```
01_pipes/              # анонимный pipe через shell: ./peer | go run .
02_pipes_zc/           #   + vmsplice(SPLICE_F_GIFT) — zero-copy на TX
03_mkfifo/             # named pipe (FIFO в ФС) — не нужен общий родитель
04_mkfifo_zc/          #   + vmsplice через FIFO
05_sockets/            # TCP loopback (127.0.0.1)
06_sockets_zc/         #   + SO_ZEROCOPY + send(MSG_ZEROCOPY) + error-queue
07_unix_sockets/       # AF_UNIX stream — те же API, без сетевого стека
08_unix_sockets_zc/    #   + SCM_RIGHTS + memfd — передача fd, mmap у получателя
09_queue/              # POSIX message queue (mq_*)
10_shmem/              # shared memory (shm_open + mmap), busy-wait синхронизация
11_semaphore/          # shmem + POSIX-семафоры (sleeping wait вместо busy-wait)
```

## Уровни zero-copy

| Пример | Кто избегает копии | Что реально происходит |
|---|---|---|
| `02_pipes_zc`, `04_mkfifo_zc` | peer (TX) | `vmsplice(SPLICE_F_GIFT)` дарит user-страницы ядру; reader всё равно копирует |
| `06_sockets_zc` | peer (TX) | `send(MSG_ZEROCOPY)` пинит user-страницы; ядро кладёт их в TCP-очередь без `skb_copy` |
| `08_unix_sockets_zc` | **оба** | memfd + SCM_RIGHTS + mmap — данные ни разу не копируются |
| `10_shmem`, `11_semaphore` | **оба** | общий объект через named shm_open, оба mmap'ят |

Полный zero-copy end-to-end дают только последние две группы. В
остальных zc-вариантах одна из сторон всё равно читает в свой буфер.

## Когда zero-copy не помогает

- **Мелкие сообщения** (< 16 KB): оверхед на pinning страниц,
  error-queue, управление ring-буфером съедает выигрыш. На 256 байт
  `pipe_zc` примерно как `pipe`, а `tcp_zc` даже чуть медленнее.
- **Короткие сессии**: setup тяжелее (memfd_create, SCM_RIGHTS,
  sockopt) — для 1–2 сообщений быстрее обычный `write`.
- **Задача не "передача байтов"**: если данные надо тут же обработать
  (парсить, считать чексумму), они всё равно попадут в регистры CPU
  и кеш. Copy-free-IPC не отменяет copy-into-registers.

См. `../benchmark/` — там видно, где какой вариант реально выигрывает.

## Общее про сигналы, shmem и lifecycle

- `signal.Notify` в Go — подписка через канал; требует буферизованный
  канал иначе сигнал теряется. См. `../signals/graceful-shutdown/`.
- Named-объекты (`/dev/shm/*`, `/dev/mqueue/*`, `/tmp/*`) переживают
  процесс, если не `unlink`'ать. Делайте `unlink` **до** `create` в
  начале демо — защита от остатков прошлого запуска.
- POSIX message queue: дефолтный системный лимит `msgsize_max=8192`
  (`/proc/sys/fs/mqueue/msgsize_max`). Для больших сообщений —
  `shmem/semaphore` или `SCM_RIGHTS+memfd`.
